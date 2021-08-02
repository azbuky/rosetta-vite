package vite

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	viteTypes "github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

func ConstructionPreprocess(
	operations []*types.Operation,
	metadata map[string]interface{},
) (*ConstructionOptions, []*types.AccountIdentifier, error) {

	matches, matched := MatchRequestOperationType(operations)

	isReceiveType := false
	if !matched {
		matches, matched = MatchResponseOperationType(operations)
		isReceiveType = true
		if !matched {
			return nil, nil, fmt.Errorf("failed to match operations")
		}
	}

	err := ValidateMatches(matches, nil)

	if err != nil {
		return nil, nil, err
	}

	fromOp, _ := matches[0].First()
	toOp, _ := matches[1].First()
	account := *fromOp.Account
	amount := *fromOp.Amount

	if isReceiveType {
		account = *toOp.Account
		amount = *toOp.Amount
	}

	// Defaults to true
	fetchPreviousHash := strconv.FormatBool(true)

	// Defaults to false
	usePow := strconv.FormatBool(false)
	usePowMeta := metadata["use_pow"]
	if usePowMeta != nil {
		usePowStr := usePowMeta.(string)
		usePowBool, err := strconv.ParseBool(usePowStr)
		if err == nil {
			usePow = strconv.FormatBool(usePowBool)
		}
	}

	var data *string
	dataMeta := metadata["data"]
	if dataMeta != nil {
		dataStr := dataMeta.(string)
		data = &dataStr
	}

	options := &ConstructionOptions{
		AccountIdentifier:  account,
		FromAccount:        *fromOp.Account,
		ToAccount:          *toOp.Account,
		Amount:             amount,
		OperationType:      fromOp.Type,
		FetchPreviousBlock: fetchPreviousHash,
		UsePow:             usePow,
		Data:               data,
	}

	requiredPublicKeys := []*types.AccountIdentifier{
		&account,
	}

	return options, requiredPublicKeys, nil
}

func (ec *Client) MatchSendBlock(ctx context.Context, options ConstructionOptions) (string, error) {
	address, err := viteTypes.HexToAddress(options.AccountIdentifier.Address)
	if err != nil {
		return "", err
	}

	pageIndex := uint64(0)
	pageSize := uint64(10)
	for {
		blocks, err := ec.c.GetUnreceivedBlocksByAddress(ctx, address, pageIndex, pageSize)
		if err != nil {
			return "", err
		}
		if len(blocks) == 0 {
			return "", fmt.Errorf("no match found")
		}
		for _, block := range blocks {
			tx, err := AccountBlockToTransaction(block, false)
			if err != nil {
				continue
			}
			match, matched := MatchRequestTransaction(tx.Operations)
			if !matched {
				continue
			}

			fromValue, err := types.AmountValue(&match.Amount)
			if err != nil {
				continue
			}
			value, err := types.AmountValue(&options.Amount)
			if err != nil {
				continue
			}

			if match.FromAccount.Address == options.FromAccount.Address &&
				match.ToAccount.Address == options.ToAccount.Address &&
				match.Amount.Currency.Metadata["tti"] == options.Amount.Currency.Metadata["tti"] &&
				fromValue != nil && value != nil &&
				fromValue.CmpAbs(value) == 0 {
				// found a match
				return block.Hash.Hex(), nil
			}

		}
	}
}

func (ec *Client) ConstructionMetadata(
	ctx context.Context,
	options *ConstructionOptions,
) (*ConstructionMetadata, error) {
	address, err := viteTypes.HexToAddress(options.AccountIdentifier.Address)
	if err != nil {
		return nil, err
	}

	accountBlock, err := ec.c.GetLatestAccountBlock(ctx, address)
	if err != nil {
		return nil, err
	}

	hash := accountBlock.Hash
	var height uint64 = 0
	if len(accountBlock.Height) > 0 {
		height, err = strconv.ParseUint(accountBlock.Height, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	metadata := &ConstructionMetadata{
		Height:       height,
		PreviousHash: accountBlock.Hash.Hex(),
	}

	if IsReceiveTypeOperation(options.OperationType) {
		sendBlockHash, err := ec.MatchSendBlock(ctx, *options)
		if err != nil {
			return nil, err
		}
		metadata.SendBlockHash = &sendBlockHash
	}

	usePow, err := strconv.ParseBool(options.UsePow)
	if err == nil && usePow {
		toAddress, err := viteTypes.HexToAddress(options.ToAccount.Address)
		if err != nil {
			return nil, err
		}

		blockType, err := OperationTypeToBlockType(options.OperationType)
		if err != nil {
			return nil, err
		}

		param := &api.GetPoWDifficultyParam{
			SelfAddr:  address,
			PrevHash:  hash,
			BlockType: blockType,
			ToAddr:    &toAddress,
		}

		if options.Data != nil {
			param.Data = []byte(*options.Data)
		}

		result, err := ec.c.GetPoWDifficulty(ctx, param)
		if err != nil {
			return nil, err
		}

		metadata.Difficulty = &result.Difficulty
		nonceHash := viteTypes.DataHash(append(address.Bytes(), hash.Bytes()...))
		nonce, err := ec.c.GetPoWNonce(ctx, result.Difficulty, nonceHash.Hex())
		if err != nil {
			return nil, err
		}

		metadata.Nonce = &nonce
	}

	return metadata, nil
}

func CreateAccountBlock(
	description *TransactionDescription,
	metadata *ConstructionMetadata,
	publicKey *types.PublicKey,
) (*api.AccountBlock, error) {
	fromAdd := description.FromAccount.Address
	toAdd := description.ToAccount.Address

	address, err := viteTypes.HexToAddress(description.Account.Address)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid address", description.Account.Address)
	}

	// Ensure valid from address
	checkFrom, err := viteTypes.HexToAddress(fromAdd)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid address", fromAdd)
	}

	// Ensure valid to address
	checkTo, err := viteTypes.HexToAddress(toAdd)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid address", toAdd)
	}

	// Ensure valid previous hash
	prevHash, err := viteTypes.HexToHash(metadata.PreviousHash)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid hash", prevHash)
	}

	blockType, err := OperationTypeToBlockType(description.OperationType)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid operation type", description.OperationType)
	}

	accountBlock := &api.AccountBlock{
		BlockType:    blockType,
		Height:       strconv.FormatUint(metadata.Height+1, 10),
		PreviousHash: prevHash,
		PrevHash:     prevHash,
		Address:      address,
		FromAddress:  checkFrom,
		ToAddress:    checkTo,
		PublicKey:    publicKey.Bytes,
	}

	if metadata.Data != nil {
		accountBlock.Data = []byte(*metadata.Data)
	}
	if metadata.Difficulty != nil {
		accountBlock.Difficulty = metadata.Difficulty
	}
	if metadata.Nonce != nil {
		accountBlock.Nonce, err = base64.StdEncoding.DecodeString(*metadata.Nonce)
		if err != nil {
			return nil, err
		}
	}

	//if IsSendTypeOperation(description.OperationType) {
	tti := description.Amount.Currency.Metadata["tti"]
	if tti == nil {
		return nil, fmt.Errorf("missing token type id")
	}
	tokenId, err := viteTypes.HexToTokenTypeId(tti.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid token type id")
	}
	accountBlock.TokenId = tokenId
	accountBlock.TokenInfo = &api.RpcTokenInfo{
		TokenSymbol: description.Amount.Currency.Symbol,
		Decimals:    uint8(description.Amount.Currency.Decimals),
		TokenId:     tokenId,
	}
	accountBlock.Amount = &description.Amount.Value
	//}
	if IsReceiveTypeOperation(description.OperationType) {
		if metadata.SendBlockHash == nil {
			return nil, fmt.Errorf("missing SendBlockHash")
		}
		sendBlockHash, err := viteTypes.HexToHash(*metadata.SendBlockHash)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid hash", *metadata.SendBlockHash)
		}
		accountBlock.SendBlockHash = sendBlockHash
		accountBlock.FromBlockHash = sendBlockHash
	}

	if description.Fee != nil {
		accountBlock.Fee = &description.Fee.Value
	}

	return accountBlock, nil
}
