package services

import (
	"context"

	"github.com/azbuky/rosetta-vite/vite"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

// Client is used by the services to get block
// data and to submit transactions.
type Client interface {
	Status(context.Context) (
		*types.BlockIdentifier,
		int64,
		*types.SyncStatus,
		[]*types.Peer,
		error,
	)

	GenesisBlockIdentifier() *types.BlockIdentifier

	Block(
		context.Context,
		*types.PartialBlockIdentifier,
	) (*types.Block, []*types.TransactionIdentifier, error)

	BlockTransaction(
		context.Context,
		*types.BlockTransactionRequest,
	) (*types.Transaction, error)

	Balance(
		context.Context,
		*types.AccountIdentifier,
		[]*types.Currency,
		*types.PartialBlockIdentifier,
	) (*types.AccountBalanceResponse, error)

	ConstructionMetadata(
		context.Context,
		*vite.ConstructionOptions,
	) (*vite.ConstructionMetadata, error)

	SendTransaction(context.Context, *api.AccountBlock) error
}
