package vite

import (
	"crypto/ed25519"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

func ConvertSecondsToMiliseconds(time int64) int64 {
	return time * 1000
}

func IsSendTypeOperation(opType string) bool {
	blockType, err := OperationTypeToBlockType(opType)
	if err != nil {
		return false
	}
	return ledger.IsSendBlock(blockType)
}

func IsReceiveTypeOperation(opType string) bool {
	blockType, err := OperationTypeToBlockType(opType)
	if err != nil {
		return false
	}
	return ledger.IsReceiveBlock(blockType)
}

func OperationTypeToBlockType(opType string) (byte, error) {
	switch opType {
	case CreateContractOpType:
		return ledger.BlockTypeSendCreate, nil
	case RequestOpType:
		return ledger.BlockTypeSendCall, nil
	case MintOpType:
		return ledger.BlockTypeSendReward, nil
	case ResponseOpType:
		return ledger.BlockTypeReceive, nil
	case ResponseFailOpType:
		return ledger.BlockTypeReceiveError, nil
	case RefundOpType:
		return ledger.BlockTypeSendRefund, nil
	case GenesisOpType:
		return ledger.BlockTypeGenesisReceive, nil
	case FeeOpType, BurnOpType:
		return 0, fmt.Errorf("op %s does not map to a block type", opType)
	default:
		return 0, fmt.Errorf("unknown operation type %s", opType)
	}
}

func BlockTypeToOperationType(blockType byte) (string, error) {
	switch blockType {
	case ledger.BlockTypeSendCreate:
		return CreateContractOpType, nil
	case ledger.BlockTypeSendCall:
		return RequestOpType, nil
	case ledger.BlockTypeSendReward:
		return MintOpType, nil
	case ledger.BlockTypeReceive:
		return ResponseOpType, nil
	case ledger.BlockTypeReceiveError:
		return ResponseFailOpType, nil
	case ledger.BlockTypeSendRefund:
		return RefundOpType, nil
	case ledger.BlockTypeGenesisReceive:
		return GenesisOpType, nil
	default:
		return "", fmt.Errorf("unknown block type %d", blockType)
	}
}

func ViteTokenToCurrency(tti string, tokenInfo api.RpcTokenInfo) (currency types.Currency) {
	currency = types.Currency{
		Symbol:   tokenInfo.TokenSymbol,
		Decimals: int32(tokenInfo.Decimals),
		Metadata: map[string]interface{}{
			"tti": tti,
		},
	}
	return
}

func CurrencyForAccountBlock(account *api.AccountBlock) *types.Currency {
	tokenInfo := account.TokenInfo
	// Workaround for token without symbol
	symbol := account.TokenId.Hex()
	decimals := int32(0)
	if tokenInfo != nil {
		symbol = tokenInfo.TokenSymbol
		decimals = int32(tokenInfo.Decimals)
	}
	return &types.Currency{
		Symbol:   symbol,
		Decimals: decimals,
		Metadata: map[string]interface{}{
			"tti": account.TokenId.Hex(),
		},
	}
}

func AmountForAccountBlock(account *api.AccountBlock, negateValue bool) *types.Amount {
	value := "0"
	if account.Amount != nil {
		value = *account.Amount
	}
	if negateValue && value != "0" {
		value = "-" + value
	}

	currency := CurrencyForAccountBlock(account)

	return &types.Amount{
		Value:    value,
		Currency: currency,
	}
}

func FeeAmountForAccountBlock(accountBlock *api.AccountBlock) *types.Amount {
	if accountBlock.Fee == nil || *accountBlock.Fee == "0" {
		return nil
	}

	currency := CurrencyForAccountBlock(accountBlock)
	return &types.Amount{
		Value:    "-" + *accountBlock.Fee,
		Currency: currency,
	}
}

func StatusRef(status string, includeStatus bool) *string {
	if !includeStatus {
		return nil
	}
	return &status
}

func FromOperationForAccountBlock(accountBlock *api.AccountBlock, index int64, includeStatus bool) (*types.Operation, error) {
	if !ledger.IsSendBlock(accountBlock.BlockType) {
		return nil, fmt.Errorf("incorrect account block type")
	}

	opType, err := BlockTypeToOperationType(accountBlock.BlockType)
	if err != nil {
		return nil, err
	}

	amount := AmountForAccountBlock(accountBlock, true)
	if opType == MintOpType {
		amount = nil
	}

	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: index,
		},
		Type:   opType,
		Status: StatusRef(SuccessStatus, includeStatus),
		Account: &types.AccountIdentifier{
			Address: accountBlock.FromAddress.Hex(),
		},
		Amount: amount,
	}, nil
}

func ToOperationForAccountBlock(accountBlock *api.AccountBlock, index int64, includeStatus bool) (*types.Operation, error) {
	opType := ResponseOpType

	amount := AmountForAccountBlock(accountBlock, false)

	if accountBlock.ToAddress.Hex() == MintAddress {
		opType = BurnOpType
		amount = nil
	}

	status := StatusRef(SuccessStatus, includeStatus)
	if ledger.IsSendBlock(accountBlock.BlockType) {
		status = StatusRef(IntentStatus, includeStatus)
	}

	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: index,
		},
		Type:   opType,
		Status: status,
		Account: &types.AccountIdentifier{
			Address: accountBlock.ToAddress.Hex(),
		},
		Amount: amount,
	}, nil
}

func FeeOperationForAccountBlock(accountBlock *api.AccountBlock, index int64, includeStatus bool) (*types.Operation, error) {
	amount := FeeAmountForAccountBlock(accountBlock)
	if amount == nil {
		return nil, fmt.Errorf("could not create Fee operation, fee value is 0 or missing")
	}

	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: index,
		},
		Type:   FeeOpType,
		Status: StatusRef(SuccessStatus, includeStatus),
		Account: &types.AccountIdentifier{
			Address: accountBlock.FromAddress.Hex(),
		},
		Amount: amount,
	}, nil
}

func OperationsForRequestAccountBlock(accountBlock *api.AccountBlock, includeStatus bool) ([]*types.Operation, error) {
	if !ledger.IsSendBlock(accountBlock.BlockType) {
		return nil, fmt.Errorf("incorrect account block type")
	}

	ops := []*types.Operation{}

	fromOp, err := FromOperationForAccountBlock(accountBlock, 0, includeStatus)
	if err != nil {
		return nil, err
	}

	toOp, err := ToOperationForAccountBlock(accountBlock, 1, includeStatus)
	if err != nil {
		return nil, err
	}
	toOp.RelatedOperations = []*types.OperationIdentifier{
		{
			Index: 0,
		},
	}
	ops = append(ops, fromOp, toOp)

	feeOp, _ := FeeOperationForAccountBlock(accountBlock, 2, includeStatus)
	if feeOp != nil {
		ops = append(ops, feeOp)
	}

	return ops, nil
}

func OperationsForResponseAccountBlock(accountBlock *api.AccountBlock, includeStatus bool) ([]*types.Operation, error) {
	if !ledger.IsReceiveBlock(accountBlock.BlockType) {
		return nil, fmt.Errorf("incorrect account block type")
	}

	ops := []*types.Operation{}

	toOp, err := ToOperationForAccountBlock(accountBlock, 0, includeStatus)
	if err != nil {
		return nil, err
	}

	ops = append(ops, toOp)

	if accountBlock.SendBlockList != nil {
		for _, sendAccount := range accountBlock.SendBlockList {
			if sendAccount == nil {
				continue
			}
			sOp, err := FromOperationForAccountBlock(sendAccount, int64(len(ops)), includeStatus)
			if err != nil {
				return nil, err
			}

			ops = append(ops, sOp)
		}
	}

	return ops, nil
}

func RelatedTransactionsForAccountBlock(accountBlock *api.AccountBlock) []*types.RelatedTransaction {
	if !ledger.IsReceiveBlock(accountBlock.BlockType) {
		return nil
	}

	rts := []*types.RelatedTransaction{
		{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: accountBlock.SendBlockHash.Hex(),
			},
			Direction: types.Backward,
		},
	}

	if accountBlock.SendBlockList != nil {
		for _, sendAccount := range accountBlock.SendBlockList {
			if sendAccount == nil {
				continue
			}
			rt := &types.RelatedTransaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: sendAccount.Hash.Hex(),
				},
				Direction: types.Forward,
			}
			rts = append(rts, rt)
		}
	}

	return rts
}

func OperationsForAccountBlock(accountBlock *api.AccountBlock, includeStatus bool) ([]*types.Operation, error) {
	if ledger.IsSendBlock(accountBlock.BlockType) {
		return OperationsForRequestAccountBlock(accountBlock, includeStatus)
	} else {
		return OperationsForResponseAccountBlock(accountBlock, includeStatus)
	}
}

// Converts a vite account block to a rosetta transaction
func AccountBlockToTransaction(accountBlock *api.AccountBlock, includeStatus bool) (*types.Transaction, error) {
	ops, err := OperationsForAccountBlock(accountBlock, includeStatus)
	if err != nil {
		return nil, err
	}

	relatedTransactions := RelatedTransactionsForAccountBlock(accountBlock)

	publicKey := ed25519.PublicKey(accountBlock.PublicKey)

	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: accountBlock.Hash.Hex(),
		},
		Operations:          ops,
		RelatedTransactions: relatedTransactions,
		Metadata: map[string]interface{}{
			"blockType":           accountBlock.BlockType,
			"height":              accountBlock.Height,
			"previousHash":        accountBlock.PrevHash,
			"publicKey":           publicKey,
			"producer":            accountBlock.Producer.Hex(),
			"fromAddress":         accountBlock.FromAddress.Hex(),
			"toAddress":           accountBlock.ToAddress.Hex(),
			"sendBlockHash":       accountBlock.SendBlockHash,
			"vmLogHash":           accountBlock.VmLogHash,
			"sendBlockList":       accountBlock.SendBlockList,
			"fee":                 accountBlock.Fee,
			"data":                accountBlock.Data,
			"difficulty":          accountBlock.Difficulty,
			"nonce":               accountBlock.Nonce,
			"signature":           accountBlock.Signature,
			"quotaUsed":           accountBlock.QuotaUsed,
			"firstSnapshotHash":   accountBlock.FirstSnapshotHash,
			"firstSnapshotHeight": accountBlock.FirstSnapshotHeight,
			// timestamp is in miliseconds
			"timestamp": ConvertSecondsToMiliseconds(accountBlock.Timestamp),
		},
	}, nil
}
