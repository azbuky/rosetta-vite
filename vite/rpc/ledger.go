package rpc

import (
	"context"

	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/rpc"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

// LedgerApi ...
type LedgerApi interface {
	GetSnapshotGenesisBlock() (*api.SnapshotBlock, error)

	GetSnapshotBlockByHash(ctx context.Context, hash types.Hash) (*api.SnapshotBlock, error)
	GetSnapshotBlockByHeight(ctx context.Context, height uint64) (*api.SnapshotBlock, error)
	GetLatestSnapshotHash(ctx context.Context) (*types.Hash, error)

	GetAccountBlockByHash(ctx context.Context, blockHash types.Hash) (*api.AccountBlock, error)
	GetAccountBlocks(ctx context.Context, address types.Address, hash *types.Hash, count uint64) ([]*api.AccountBlock, error)
	GetAccountInfoByAddress(ctx context.Context, address types.Address) (*api.AccountInfo, error)
	GetConfirmedBalances(ctx context.Context, snapshotHash types.Hash, addrList []types.Address, tokenIds []types.TokenTypeId) (result *api.GetBalancesRes, err error)
	GetLatestAccountBlock(ctx context.Context, address types.Address) (*api.AccountBlock, error)
	GetUnreceivedBlocksByAddress(ctx context.Context, address types.Address, page uint64, pageSize uint64) ([]*api.AccountBlock, error)

	GetPoWDifficulty(ctx context.Context, param *api.GetPoWDifficultyParam) (*api.GetPoWDifficultyResult, error)

	SendRawTransaction(ctx context.Context, accountBlock *api.AccountBlock) error
}

type ledgerApi struct {
	cc *rpc.Client
}

func NewLedgerApi(cc *rpc.Client) LedgerApi {
	return &ledgerApi{cc: cc}
}

func (li ledgerApi) GetAccountBlockByHash(
	ctx context.Context,
	hash types.Hash,
) (block *api.AccountBlock, err error) {
	block = &api.AccountBlock{}
	err = li.cc.CallContext(ctx, block, "ledger_getBlockByHash", hash)
	return
}

func (li ledgerApi) GetAccountBlocks(
	ctx context.Context,
	address types.Address,
	hash *types.Hash,
	count uint64,
) (blocks []*api.AccountBlock, err error) {
	blocks = []*api.AccountBlock{}
	err = li.cc.CallContext(ctx, blocks, "ledger_getAccountBlocks", address, hash, nil, count)
	return
}

func (li ledgerApi) GetAccountInfoByAddress(
	ctx context.Context,
	address types.Address,
) (accountInfo *api.AccountInfo, err error) {
	accountInfo = &api.AccountInfo{}
	err = li.cc.CallContext(ctx, accountInfo, "ledger_getAccountInfoByAddress", address)
	return
}

func (li ledgerApi) GetSnapshotGenesisBlock() (block *api.SnapshotBlock, err error) {
	block = &api.SnapshotBlock{}
	err = li.cc.Call(block, "ledger_getSnapshotBlockByHeight", 1)
	return
}

func (li ledgerApi) GetLatestAccountBlock(
	ctx context.Context,
	address types.Address,
) (accountBlock *api.AccountBlock, err error) {
	accountBlock = &api.AccountBlock{}
	err = li.cc.CallContext(ctx, accountBlock, "ledger_getLatestAccountBlock", address)
	return
}

func (li ledgerApi) GetUnreceivedBlocksByAddress(
	ctx context.Context,
	address types.Address,
	page uint64,
	pageSize uint64,
) (result []*api.AccountBlock, err error) {
	result = []*api.AccountBlock{}
	err = li.cc.CallContext(ctx, &result, "ledger_getUnreceivedBlocksByAddress", address, page, pageSize)
	return
}

func (li ledgerApi) GetPoWDifficulty(
	ctx context.Context,
	param *api.GetPoWDifficultyParam,
) (result *api.GetPoWDifficultyResult, err error) {
	result = &api.GetPoWDifficultyResult{}
	err = li.cc.CallContext(ctx, result, "ledger_getPoWDifficulty", param)
	return
}

func (li ledgerApi) SendRawTransaction(
	ctx context.Context,
	accountBlock *api.AccountBlock,
) error {
	err := li.cc.CallContext(ctx, nil, "ledger_sendRawTransaction", accountBlock)
	return err
}

func (li ledgerApi) GetSnapshotBlockByHash(ctx context.Context, hash types.Hash) (block *api.SnapshotBlock, err error) {
	block = &api.SnapshotBlock{}
	err = li.cc.CallContext(ctx, block, "ledger_getSnapshotBlockByHash", hash)
	return
}

func (li ledgerApi) GetSnapshotBlockByHeight(ctx context.Context, height uint64) (block *api.SnapshotBlock, err error) {
	block = &api.SnapshotBlock{}
	err = li.cc.CallContext(ctx, block, "ledger_getSnapshotBlockByHeight", height)
	return
}

func (li ledgerApi) GetLatestSnapshotHash(ctx context.Context) (hash *types.Hash, err error) {
	hash = &types.Hash{}
	err = li.cc.CallContext(ctx, hash, "ledger_getLatestSnapshotHash")
	return
}

func (li ledgerApi) GetConfirmedBalances(
	ctx context.Context,
	snapshotHash types.Hash,
	addrList []types.Address,
	tokenIds []types.TokenTypeId,
) (result *api.GetBalancesRes, err error) {
	result = &api.GetBalancesRes{}
	err = li.cc.CallContext(ctx, &result, "ledger_getConfirmedBalances", snapshotHash, addrList, tokenIds)
	return
}
