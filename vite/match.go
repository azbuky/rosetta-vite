package vite

import (
	"fmt"
	"reflect"

	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	viteTypes "github.com/vitelabs/go-vite/common/types"
)

const SendBlockHashKey string = "sendBlockHash"
const DataKey string = "data"

func MatchTransaction(operations []*types.Operation) (*TransactionDescription, error) {
	if len(operations) == 0 {
		return nil, fmt.Errorf("missing operations")
	}

	description, err := MatchRequestTransaction(operations)
	if err == nil {
		return description, nil
	}

	description, err = MatchResponseTransaction(operations)
	if err == nil {
		return description, nil
	}

	return nil, fmt.Errorf("could not match operations")
}

func MatchRequestTransaction(operations []*types.Operation) (*TransactionDescription, error) {
	if len(operations) != 2 {
		return nil, fmt.Errorf("incorrect number of ops")
	}

	reqOp := operations[0]
	respOp := operations[1]

	if err := CheckRequestOpType(reqOp); err != nil {
		return nil, err
	}

	if err := CheckResponseOpType(respOp, false); err != nil {
		return nil, err
	}

	// convert amount to positive value
	amount := *reqOp.Amount
	value, err := types.NegateValue(amount.Value)
	if err != nil {
		return nil, err
	}
	amount.Value = value

	// check for data field
	var data *string
	dataField, ok := reqOp.Metadata[DataKey].(string)
	if ok {
		data = &dataField
	}

	transaction := &TransactionDescription{
		OperationType: RequestOpType,
		Account:       *reqOp.Account,
		FromAccount:   reqOp.Account,
		ToAccount:     *respOp.Account,
		Amount:        amount,
		Data:          data,
	}
	return transaction, nil
}

func MatchResponseTransaction(operations []*types.Operation) (*TransactionDescription, error) {
	if len(operations) != 1 {
		return nil, fmt.Errorf("incorrect number of ops")
	}

	respOp := operations[0]

	if err := CheckResponseOpType(respOp, true); err != nil {
		return nil, err
	}

	// check for data field
	var data *string
	dataField, ok := respOp.Metadata[DataKey].(string)
	if ok {
		data = &dataField
	}

	sendBlockHash := &types.TransactionIdentifier{
		Hash: respOp.Metadata[SendBlockHashKey].(string),
	}

	transaction := &TransactionDescription{
		OperationType: ResponseOpType,
		Account:       *respOp.Account,
		ToAccount:     *respOp.Account,
		Amount:        *respOp.Amount,
		SendBlockHash: sendBlockHash,
		Data:          data,
	}

	return transaction, nil
}

func CheckRequestOpType(operation *types.Operation) error {
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
		},
		ErrUnmatched: true,
	}

	match, err := parser.MatchOperations(description, []*types.Operation{operation})
	if err != nil {
		return err
	}

	if err = ValidateMatch(match[0]); err != nil {
		return err
	}

	return nil
}

func CheckResponseOpType(operation *types.Operation, inResponseTx bool) error {
	var metadata []*parser.MetadataDescription
	if inResponseTx {
		metadata = []*parser.MetadataDescription{
			{
				Key:       SendBlockHashKey,
				ValueKind: reflect.String,
			},
		}
	}

	description := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: ResponseOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: true,
					Sign:   parser.PositiveOrZeroAmountSign,
				},
				Metadata: metadata,
			},
		},
		ErrUnmatched: true,
	}

	match, err := parser.MatchOperations(description, []*types.Operation{operation})
	if err != nil {
		return err
	}

	if err = ValidateMatch(match[0]); err != nil {
		return err
	}

	return nil
}

func CheckBurnOpType(operation *types.Operation) error {
	description := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type: BurnOpType,
				Account: &parser.AccountDescription{
					Exists: true,
				},
				Amount: &parser.AmountDescription{
					Exists: false,
				},
			},
		},
		ErrUnmatched: true,
	}

	match, err := parser.MatchOperations(description, []*types.Operation{operation})
	if err != nil {
		return err
	}

	if err = ValidateMatch(match[0]); err != nil {
		return err
	}

	// check that the op address is the mint address
	op, _ := match[0].First()
	if op.Account.Address != MintAddress {
		return fmt.Errorf("incorrect address for burn op")
	}

	return nil
}

func CheckFeeOpType(operation *types.Operation) error {
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

	match, err := parser.MatchOperations(description, []*types.Operation{operation})
	if err != nil {
		return err
	}

	if err := ValidateMatch(match[0]); err != nil {
		return err
	}

	return nil
}

func ValidateMatch(match *parser.Match) error {
	op, _ := match.First()

	address := op.Account.Address
	// Ensure valid account address
	_, err := viteTypes.HexToAddress(address)
	if err != nil {
		return fmt.Errorf("%s is not a valid address", address)
	}

	return nil
}
