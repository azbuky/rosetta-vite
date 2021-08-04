package vite

import (
	"crypto/ed25519"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	viteTypes "github.com/vitelabs/go-vite/common/types"
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

func AmountForAccountBlock(account *api.AccountBlock) *types.Amount {
	value := "0"
	if account.Amount != nil {
		value = *account.Amount
	}
	if ledger.IsSendBlock(account.BlockType) && value != "0" {
		value = "-" + value
	}

	currency := CurrencyForAccountBlock(account)

	return &types.Amount{
		Value:    value,
		Currency: currency,
	}
}

func OperationsForAccountBlock(account *api.AccountBlock, startIndex int64, includeStatus bool) ([]*types.Operation, error) {
	ops := []*types.Operation{}

	opType, err := BlockTypeToOperationType(account.BlockType)
	if err != nil {
		return nil, err
	}

	currency := CurrencyForAccountBlock(account)
	zeroAmount := &types.Amount{
		Value:    "0",
		Currency: currency,
	}

	amount := AmountForAccountBlock(account)
	sStatus := SuccessStatus
	var status *string
	if includeStatus {
		status = &sStatus
	}

	fromOp := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: startIndex,
		},
		Type:   opType,
		Status: status,
		Account: &types.AccountIdentifier{
			Address: account.FromAddress.Hex(),
		},
		Amount: amount,
	}

	toOp := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: startIndex + 1,
		},
		RelatedOperations: []*types.OperationIdentifier{
			{
				Index: startIndex,
			},
		},
		Type:   opType,
		Status: status,
		Account: &types.AccountIdentifier{
			Address: account.ToAddress.Hex(),
		},
		Amount: amount,
	}

	if ledger.IsSendBlock(account.BlockType) {
		toOp.Amount = zeroAmount
		ops = append(ops, fromOp, toOp)
		// only include free if not zero
		if account.Fee != nil && *account.Fee != "0" {
			fee := *account.Fee
			feeOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: startIndex + int64(len(ops)),
				},
				RelatedOperations: []*types.OperationIdentifier{
					{
						Index: startIndex,
					},
				},
				Type:   FeeOpType,
				Status: status,
				Account: &types.AccountIdentifier{
					Address: account.FromAddress.Hex(),
				},
				Amount: &types.Amount{
					Value:    "-" + fee,
					Currency: currency,
				},
			}
			ops = append(ops, feeOp)
		}
	} else {
		fromOp.Amount = zeroAmount
		ops = append(ops, fromOp, toOp)

		// handle inline send block transactions
		if account.SendBlockList != nil {
			for _, sendAccount := range account.SendBlockList {
				if sendAccount == nil {
					continue
				}
				sOps, err := OperationsForAccountBlock(sendAccount, int64(len(ops)), includeStatus)
				if err != nil {
					return nil, err
				}
				// Do not substract reward amount from mint address
				if sendAccount.BlockType == ledger.BlockTypeSendReward {
					sOps[0].Amount.Value = "0"
				}
				ops = append(ops, sOps...)
			}
		}
	}

	return ops, nil
}

// Converts a vite account block to a rosetta transaction
func AccountBlockToTransaction(accountBlock *api.AccountBlock, includeStatus bool) (*types.Transaction, error) {
	ops, err := OperationsForAccountBlock(accountBlock, 0, includeStatus)
	if err != nil {
		return nil, err
	}

	var direction types.Direction
	var relatedHash *viteTypes.Hash
	if ledger.IsSendBlock(accountBlock.BlockType) {
		direction = types.Forward
		if accountBlock.ReceiveBlockHash != nil {
			relatedHash = accountBlock.ReceiveBlockHash
		}
	} else {
		direction = types.Backward
		relatedHash = &accountBlock.SendBlockHash
	}

	relatedTransactions := []*types.RelatedTransaction{}
	if relatedHash != nil {
		rt := &types.RelatedTransaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: relatedHash.Hex(),
			},
			Direction: direction,
		}
		relatedTransactions = append(relatedTransactions, rt)
	}

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
			"dificulty":           accountBlock.Difficulty,
			"nonce":               accountBlock.Nonce,
			"signature":           accountBlock.Signature,
			"quotaUsed":           accountBlock.QuotaUsed,
			"confirmations":       accountBlock.Confirmations,
			"firstSnapshotHash":   accountBlock.FirstSnapshotHash,
			"firstSnapshotHeight": accountBlock.FirstSnapshotHeight,
			"receiveBlockHeight":  accountBlock.ReceiveBlockHeight,
			"receiveBlockHash":    accountBlock.ReceiveBlockHash,
			// timestamp is in miliseconds
			"timestamp": ConvertSecondsToMiliseconds(accountBlock.Timestamp),
		},
	}, nil
}
