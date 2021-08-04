package vite

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	viteTypes "github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
)

func MatchTransaction(operations []*types.Operation) (*TransactionDescription, error) {
	description, matched := MatchRequestTransaction(operations)
	if matched {
		return description, nil
	}

	description, matched = MatchResponseTransaction(operations)
	if matched {
		return description, nil
	}

	return nil, fmt.Errorf("could not match operations")
}

func MatchRequestTransaction(operations []*types.Operation) (*TransactionDescription, bool) {
	matches, matched := MatchRequestOperationType(operations)
	if !matched {
		return nil, false
	}

	opType := RequestOpType
	err := ValidateMatches(matches, &opType)
	if err != nil {
		return nil, false
	}
	fromOp, fromValue := matches[0].First()
	toOp, _ := matches[1].First()

	feeMatch, matched := MatchFeeOperationType(operations)

	var fee *types.Amount
	if matched {
		feeOp, _ := feeMatch[0].First()
		fee = feeOp.Amount
	}

	// convert amount to positive value
	amount := *fromOp.Amount
	amount.Value = fromValue.Abs(fromValue).String()

	transaction := &TransactionDescription{
		OperationType: RequestOpType,
		Account:       *fromOp.Account,
		FromAccount:   *fromOp.Account,
		ToAccount:     *toOp.Account,
		Amount:        amount,
		Fee:           fee,
	}
	return transaction, true
}

func MatchResponseTransaction(operations []*types.Operation) (*TransactionDescription, bool) {
	matches, matched := MatchResponseOperationType(operations)
	if !matched {
		return nil, false
	}

	opType := ResponseOpType
	err := ValidateMatches(matches, &opType)
	if err != nil {
		return nil, false
	}

	fromOp, _ := matches[0].First()
	toOp, _ := matches[1].First()

	transaction := &TransactionDescription{
		OperationType: opType,
		Account:       *toOp.Account,
		FromAccount:   *fromOp.Account,
		ToAccount:     *toOp.Account,
		Amount:        *toOp.Amount,
	}
	return transaction, true
}

func MatchRequestOperationType(
	operations []*types.Operation,
) ([]*parser.Match, bool) {
	description := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: RequestOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.NegativeOrZeroAmountSign,
				},
			},
			{
				Type: RequestOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
				},
			},
		},
		ErrUnmatched: true,
	}

	matches, err := parser.MatchOperations(description, operations)
	if err != nil {
		return nil, false
	}

	return matches, true
}

func MatchResponseOperationType(
	operations []*types.Operation,
) ([]*parser.Match, bool) {
	description := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: ResponseOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
				},
			},
			{
				Type: ResponseOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.PositiveOrZeroAmountSign,
				},
			},
		},
		ErrUnmatched: true,
	}

	matches, err := parser.MatchOperations(description, operations)
	if err != nil {
		return nil, false
	}

	return matches, true
}

func MatchFeeOperationType(
	operations []*types.Operation,
) ([]*parser.Match, bool) {
	description := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: FeeOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.NegativeAmountSign,
				},
			},
		},
		ErrUnmatched: true,
	}

	matches, err := parser.MatchOperations(description, operations)
	if err != nil {
		return nil, false
	}
	return matches, true
}

// Basic validation for a send or receive operation type pair
func ValidateMatches(matches []*parser.Match, opType *string) error {
	fromOp, fromAmount := matches[0].First()
	fromAddressStr := fromOp.Account.Address

	toOp, toAmount := matches[1].First()
	toAddressStr := toOp.Account.Address

	// Ensure there are exactly two matches
	if len(matches) != 2 {
		return fmt.Errorf("expected 2 operations, found %d", len(matches))
	}

	if opType != nil {
		// Ensure both matches have correct operation type
		if fromOp.Type != *opType {
			return fmt.Errorf("from operation type %s does not match %s required type", toOp.Type, *opType)
		}
		if toOp.Type != *opType {
			return fmt.Errorf("to operation type %s does not match %s required type", toOp.Type, *opType)
		}
	} else {
		// Ensure both matches have same operation type
		if fromOp.Type != toOp.Type {
			return fmt.Errorf("operation types %s and %s do not match", fromOp.Type, toOp.Type)
		}
	}

	blockType, err := OperationTypeToBlockType(fromOp.Type)
	if err != nil {
		return err
	}

	isSendType := ledger.IsSendBlock(blockType)

	// Ensure valid from address
	_, err = viteTypes.HexToAddress(fromAddressStr)
	if err != nil {
		return fmt.Errorf("%s is not a valid address", fromAddressStr)
	}

	// Ensure valid to address
	_, err = viteTypes.HexToAddress(toAddressStr)
	if err != nil {
		return fmt.Errorf("%s is not a valid address", toAddressStr)
	}

	if isSendType {
		// Ensure destination amount is zero for send block types
		if len(toAmount.Bits()) != 0 {
			return fmt.Errorf("%s is not zero", toAmount.String())
		}
	} else {
		// Ensure source amount is zero for receive block types
		if len(fromAmount.Bits()) != 0 {
			return fmt.Errorf("%s is not zero", fromAmount.String())
		}
	}

	return nil
}
