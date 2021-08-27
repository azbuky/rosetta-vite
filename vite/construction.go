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

	description, err := MatchTransaction(operations)

	if err != nil {
		return nil, nil, err
	}

	// Defaults to true
	fetchPreviousHash := strconv.FormatBool(true)

	// Defaults to false
	usePow := strconv.FormatBool(false)
	usePowStr, ok := metadata["use_pow"].(string)
	if ok {
		usePowBool, err := strconv.ParseBool(usePowStr)
		if err == nil {
			usePow = strconv.FormatBool(usePowBool)
		}
	}

	options := &ConstructionOptions{
		OperationType:      description.OperationType,
		Account:            description.Account,
		ToAccount:          description.ToAccount,
		Amount:             description.Amount,
		FetchPreviousBlock: fetchPreviousHash,
		UsePow:             usePow,
		Data:               description.Data,
	}

	requiredPublicKeys := []*types.AccountIdentifier{
		&description.Account,
	}

	return options, requiredPublicKeys, nil
}

func (ec *Client) ConstructionMetadata(
	ctx context.Context,
	options *ConstructionOptions,
) (*ConstructionMetadata, error) {
	address, err := viteTypes.HexToAddress(options.Account.Address)
	if err != nil {
		return nil, err
	}

	accountBlock, err := ec.c.GetLatestAccountBlock(ctx, address)
	if err != nil {
		return nil, err
	}

	prevHash := accountBlock.Hash
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
			PrevHash:  prevHash,
			BlockType: blockType,
			ToAddr:    &toAddress,
			Data:      options.Data,
		}

		// if options.Data != nil {
		// 	param.Data = options.Data
		// }

		result, err := ec.c.GetPoWDifficulty(ctx, param)
		if err != nil {
			return nil, err
		}

		if len(result.Difficulty) > 0 {
			nonceHash := viteTypes.DataHash(append(address.Bytes(), prevHash.Bytes()...))
			nonce, err := ec.c.GetPoWNonce(ctx, result.Difficulty, nonceHash.Hex())
			if err != nil {
				return nil, err
			}

			metadata.Difficulty = &result.Difficulty
			metadata.Nonce = &nonce
		}
	}

	return metadata, nil
}

func CreateAccountBlock(
	description *TransactionDescription,
	metadata *ConstructionMetadata,
	publicKey *types.PublicKey,
) (*api.AccountBlock, error) {
	address, err := viteTypes.HexToAddress(description.Account.Address)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid address", description.Account.Address)
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
		PublicKey:    publicKey.Bytes,
		Data:         description.Data,
	}

	if description.FromAccount != nil {
		fromAdd := description.FromAccount.Address
		// Ensure valid from address
		checkFrom, err := viteTypes.HexToAddress(fromAdd)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid address", fromAdd)
		}
		accountBlock.FromAddress = checkFrom
	}

	toAdd := description.ToAccount.Address
	// Ensure valid to address
	checkTo, err := viteTypes.HexToAddress(toAdd)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid address", toAdd)
	}
	accountBlock.ToAddress = checkTo

	// if description.Data != nil {
	// 	accountBlock.Data = []byte(*description.Data)
	// }

	if metadata.Difficulty != nil {
		accountBlock.Difficulty = metadata.Difficulty
	}
	if metadata.Nonce != nil {
		accountBlock.Nonce, err = base64.StdEncoding.DecodeString(*metadata.Nonce)
		if err != nil {
			return nil, err
		}
	}

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

	if IsReceiveTypeOperation(description.OperationType) {
		if description.SendBlockHash == nil {
			return nil, fmt.Errorf("missing SendBlockHash")
		}
		sendBlockHash, err := viteTypes.HexToHash(description.SendBlockHash.Hash)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid hash", description.SendBlockHash.Hash)
		}
		accountBlock.SendBlockHash = sendBlockHash
	}

	if description.Fee != nil {
		accountBlock.Fee = &description.Fee.Value
	}

	return accountBlock, nil
}
