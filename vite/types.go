package vite

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	// NodeVersion is the version of gvite we are using.
	NodeVersion = "2.10.2"

	// Blockchain is Vite.
	Blockchain string = "vite"

	// MainnetNetwork is the value of the network
	// in MainnetNetworkIdentifier.
	MainnetNetwork string = "mainnet"

	// Testnet is the value of the network
	// in TestnetNetworkIdentifier.
	TestnetNetwork string = "testnet"

	// Devnet is the value of the network
	// in DevnetNetworkIdentifier.
	DevnetNetwork string = "devnet"

	// Symbol is the symbol value
	// used in Currency.
	Symbol = "VITE"

	// Decimals is the decimals value
	// used in Currency.
	Decimals = 18

	CreateContractOpType = "CREATE_CONTRACT"
	RequestOpType = "REQUEST"
	MintOpType = "MINT"
	ResponseOpType = "RESPONSE"
	ResponseFailOpType = "RESPONSE_FAIL"
	RefundOpType = "REFUND"
	GenesisOpType = "GENESIS"
	FeeOpType = "FEE"

	// SuccessStatus is the status of any
	// operation considered successful.
	SuccessStatus string = "SUCCESS"

	RevertedStatus string = "REVERTED"

	ExceedMaxDepthStatus string = "EXCEED_MAX_DEPTH"

	// HistoricalBalanceSupported is whether
	// historical balance is supported.
	HistoricalBalanceSupported = true

	// GenesisBlockIndex is the index of the
	// genesis block.
	GenesisBlockIndex = int64(1)

	// MainnetGviteArguments are the arguments to start a mainnet gvite instance.
	MainnetGviteArguments = `--config=/app/vite/node_config.json`

	// IncludeMempoolCoins does not apply to rosetta-vite as it is not UTXO-based.
	IncludeMempoolCoins = false

	// InlineTransactions - weather to return transactions inline in the block or
	// as otherTransactions
	InlineTransactions = true
)

var (
	// TestnetGviteArguments are the arguments to start a testnet gvite instance.
	TestnetGviteArguments = fmt.Sprintf("%s --networkid 2", MainnetGviteArguments)

	// Currency is the *types.Currency for all
	// Vite networks.
	Currency = &types.Currency{
		Symbol:   Symbol,
		Decimals: Decimals,
	}

	// OperationTypes are all suppoorted operation types.
	OperationTypes = []string{
		CreateContractOpType,
		RequestOpType,
		MintOpType,
		ResponseOpType,
		ResponseFailOpType,
		RefundOpType,
		GenesisOpType,
		FeeOpType,
	}

	// OperationStatuses are all supported operation statuses.
	OperationStatuses = []*types.OperationStatus{
		{
			Status:     SuccessStatus,
			Successful: true,
		},
		{
			Status:     RevertedStatus,
			Successful: false,
		},
		{
			Status: 	ExceedMaxDepthStatus,
			Successful: false,
		},
	}

	// CallMethods are all supported call methods.
	CallMethods = []string{}
)

// Defines construction preprocess options
type ConstructionOptions struct {
	AccountIdentifier  		types.AccountIdentifier `json:"account_identifier"`
	FromAccount 			types.AccountIdentifier `json:"from_account"`
	ToAccount				types.AccountIdentifier `json:"to_account"`
	OperationType 			string					`json:"operation_type"`
	Amount					types.Amount			`json:"amount"`
	FetchPreviousBlock 		string                  `json:"fetch_previous_block"`
	UsePow             		string                  `json:"use_pow"`
	Data					*string					`json:"data,omitempty"`
}

// Defines construction metadata
type ConstructionMetadata struct {
	Height        uint64 `json:"height"`
	PreviousHash  string `json:"previousHash"`
	Difficulty    *string `json:"difficulty,omitempty"`
	Nonce         *string `json:"nonce,omitempty"`
	SendBlockHash *string `json:"sendBlockHash,omitempty"`
	Data          *string `json:"data,omitempty"`
}

// Defines transaction description from matched operations
type TransactionDescription struct {
	OperationType 	string
	Account			types.AccountIdentifier
	FromAccount 	types.AccountIdentifier
	ToAccount		types.AccountIdentifier
	// Amount should always be positive
	Amount 			types.Amount
	Fee				*types.Amount
}
