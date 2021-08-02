package vite

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/azbuky/rosetta-vite/vite/rpc"
	"github.com/coinbase/rosetta-sdk-go/types"

	viteTypes "github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/net"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

type Client struct {
	c rpc.RpcClient

	inlineTransactions bool

	genesisBlockIdentifier *types.BlockIdentifier
}

// NewClient creates a Client that from the provided url and params.
func NewClient(url string, inlineTransactions bool) (*Client, error) {
	c, err := rpc.NewRpcClient(url)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to dial node", err)
	}

	genesisBlock, err := c.GetSnapshotGenesisBlock()
	if err != nil {
		return nil, fmt.Errorf("%w: unable to get genesis block", err)
	}

	genesisBlockIdentifier := &types.BlockIdentifier{
		Hash:  genesisBlock.Hash.Hex(),
		Index: int64(genesisBlock.Height),
	}

	return &Client{c, inlineTransactions, genesisBlockIdentifier}, nil
}

// Close shuts down the RPC client connection.
func (ec *Client) Close() {
	ec.c.GetClient().Close()
}

// GenesisBlockIdentifier returns cached genesis block identifier
func (ec *Client) GenesisBlockIdentifier() *types.BlockIdentifier {
	return ec.genesisBlockIdentifier
}

// Status returns gvite status information
// for determining node healthiness.
func (ec *Client) Status(ctx context.Context) (
	*types.BlockIdentifier,
	int64,
	*types.SyncStatus,
	[]*types.Peer,
	error,
) {
	nodeInfo, err := ec.c.GetNodeInfo(ctx)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	block, err := ec.c.GetSnapshotBlockByHeight(ctx, nodeInfo.Height)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	syncInfo, err := ec.c.GetSyncInfo(ctx)

	if err != nil {
		return nil, -1, nil, nil, err
	}

	var syncStatus *types.SyncStatus
	if syncInfo != nil {
		currentIndex, err := strconv.ParseInt(syncInfo.Current, 10, 64)
		if err != nil {
			return nil, -1, nil, nil, err
		}
		stage := fmt.Sprint(syncInfo.State)
		synced := syncInfo.State == 2

		syncStatus = &types.SyncStatus{
			CurrentIndex: &currentIndex,
			Stage:        &stage,
			Synced:       &synced,
		}
	}

	peers, err := ec.peers(nodeInfo)
	if err != nil {
		return nil, -1, nil, nil, err
	}

	blockIdentifier := ec.getBlockIdentifier(block)
	timestamp := ConvertSecondsToMiliseconds(block.Timestamp)

	return blockIdentifier,
		timestamp,
		syncStatus,
		peers,
		nil
}

// Get Peers of the node from NodeInfo.
func (ec *Client) peers(nodeInfo *net.NodeInfo) ([]*types.Peer, error) {
	peers := make([]*types.Peer, nodeInfo.PeerCount)
	for i, peerInfo := range nodeInfo.Peers {
		peers[i] = &types.Peer{
			PeerID: peerInfo.Id,
			Metadata: map[string]interface{}{
				"name":       peerInfo.Name,
				"version":    0,
				"height":     peerInfo.Height,
				"address":    peerInfo.Address,
				"flag":       0,
				"superior":   true,
				"reliable":   true,
				"createdAt":  peerInfo.CreateAt,
				"readQueue":  0,
				"writeQueue": 0,
			},
		}
	}

	return peers, nil
}

// Transaction
func (ec *Client) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.Transaction, error) {

	transactionIdentifier := request.TransactionIdentifier

	hash, err := viteTypes.HexToHash(transactionIdentifier.Hash)
	if err != nil {
		return nil, err
	}

	accountBlock, err := ec.c.GetAccountBlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	return AccountBlockToTransaction(accountBlock, true)
}

// Block returns a populated block at the *RosettaTypes.PartialBlockIdentifier.
// If neither the hash or index is populated in the blockIdentifier,
// the current block is returned.
func (ec *Client) Block(
	ctx context.Context,
	blockIdentifier *types.PartialBlockIdentifier,
) (*types.Block, []*types.TransactionIdentifier, error) {
	// get SnapshotBlock for block identifier
	block, err := ec.getSnapshotBlock(ctx, blockIdentifier)
	if err != nil {
		return nil, nil, err
	}

	// if InlineTransactions is true then parse all transactions in block
	// otherwise only get transaction ids and return them as otherTransactions
	txIds := []*types.TransactionIdentifier{}
	txs := []*types.Transaction{}

	currentIdentifier := ec.getBlockIdentifier(block)
	parentIdentifier := currentIdentifier

	// skip genesis transactions for optimisation purposes
	
	if currentIdentifier.Index != GenesisBlockIndex {
		parentIdentifier = &types.BlockIdentifier{
			Hash:  block.PreviousHash.Hex(),
			Index: currentIdentifier.Index - 1,
		}

		for _, hashHeight := range block.SnapshotData {
			hash := hashHeight.Hash
			// retrieve all account blocks for each address in selected snapshot block
			for {
				account, err := ec.c.GetAccountBlockByHash(ctx, hash)
				if err != nil {
					return nil, nil, err
				}
				if account.FirstSnapshotHash == nil || account.FirstSnapshotHash.Hex() != block.Hash.Hex() {
					break
				}
				if ec.inlineTransactions {
					transaction, err := AccountBlockToTransaction(account, true)
					if err != nil {
						return nil, nil, err
					}
					txs = append(txs, transaction)
				} else {
					txIds = append(txIds, &types.TransactionIdentifier{
						Hash: hash.Hex(),
					})
				}
				hash = account.PreviousHash
			}
		}
	}

	return &types.Block{
		BlockIdentifier:       currentIdentifier,
		ParentBlockIdentifier: parentIdentifier,
		Timestamp:             ConvertSecondsToMiliseconds(block.Timestamp),
		Transactions:          txs,
		Metadata: map[string]interface{}{
			"producer":     block.Producer,
			"publicKey":    block.PublicKey,
			"signature":    block.Signature,
			"seed":         block.Seed,
			"nextSeedHash": block.NextSeedHash,
			"version":      block.Version,
		},
	}, txIds, nil
}

// Retrieve a SnapshotBlock for the given block identifier
// if block identifier is nil, returns latest block
func (ec *Client) getSnapshotBlock(
	ctx context.Context,
	blockIdentifier *types.PartialBlockIdentifier,
) (*api.SnapshotBlock, error) {
	if blockIdentifier != nil {
		if blockIdentifier.Hash != nil {
			hash, err := viteTypes.HexToHash(*blockIdentifier.Hash)
			if err != nil {
				return nil, err
			}
			return ec.c.GetSnapshotBlockByHash(ctx, hash)
		} else if blockIdentifier.Index != nil {
			return ec.c.GetSnapshotBlockByHeight(ctx, uint64(*blockIdentifier.Index))
		}
	}

	hash, err := ec.c.GetLatestSnapshotHash(ctx)
	if err != nil {
		return nil, err
	}

	return ec.c.GetSnapshotBlockByHash(ctx, *hash)
}

// Get BlockIdentifier for a SnapshotBlock
func (ec *Client) getBlockIdentifier(
	block *api.SnapshotBlock,
) *types.BlockIdentifier {
	return &types.BlockIdentifier{
		Hash:  block.Hash.Hex(),
		Index: int64(block.Height),
	}
}

// Balance returns the balance of a vite address
// at the given snapshot block identifier
// if blockIdentifier is nil, balances for the latest snapshot block are returned
func (ec *Client) Balance(
	ctx context.Context,
	account *types.AccountIdentifier,
	currencies []*types.Currency,
	blockIdentifier *types.PartialBlockIdentifier,
) (*types.AccountBalanceResponse, error) {

	block, err := ec.getSnapshotBlock(ctx, blockIdentifier)
	if err != nil {
		return nil, err
	}

	address, err := viteTypes.HexToAddress(account.Address)
	if err != nil {
		return nil, err
	}

	tokenIds := []viteTypes.TokenTypeId{}
	for _, key := range currencies {
		ttiMeta := key.Metadata["tti"]
		if ttiMeta != nil {
			ttiStr := ttiMeta.(string)
			tti, err := viteTypes.HexToTokenTypeId(ttiStr)
			if err != nil {
				continue
			}
			tokenIds = append(tokenIds, tti)
		}

	}
	if len(currencies) == 0 {
		accountInfo, err := ec.c.GetAccountInfoByAddress(ctx, address)
		if err != nil {
			return nil, err
		}
		for key := range accountInfo.BalanceInfoMap {
			tokenIds = append(tokenIds, key)
		}
	}

	confirmedBalances, err := ec.c.GetConfirmedBalances(ctx, block.Hash, []viteTypes.Address{address}, tokenIds)
	if err != nil {
		return nil, err
	}

	// if address has no balances
	if confirmedBalances == nil {
		// create zero balances for requested currencies
		balances := []*types.Amount{}
		for _, currency := range currencies {
			balance := &types.Amount{
				Value:    "0",
				Currency: currency,
			}
			balances = append(balances, balance)
		}
		// if no currency was requested
		if len(balances) == 0 {
			// add VITE currency as default
			balance := &types.Amount{
				Value: "0",
				Currency: &types.Currency{
					Symbol:   "VITE",
					Decimals: 18,
					Metadata: map[string]interface{}{
						"tti": "tti_5649544520544f4b454e6e40",
					},
				},
			}
			balances = append(balances, balance)
		}
		return &types.AccountBalanceResponse{
			BlockIdentifier: &types.BlockIdentifier{
				Hash:  block.Hash.Hex(),
				Index: int64(block.Height),
			},
			Balances: balances,
			// Metadata: map[string]interface{}{
			// 	"address":    address.Hex(),
			// 	"blockCount": 0,
			// },
		}, nil
	}

	accountBalances := (*confirmedBalances)[address]

	balances := make([]*types.Amount, len(tokenIds))
	for i, tokenId := range tokenIds {
		tokenInfo, err := ec.c.GetTokenInfoById(ctx, tokenId.Hex())
		if err != nil {
			return nil, err
		}
		value := accountBalances[tokenId]
		if value == nil {
			value = big.NewInt(0)
		}
		currency := &types.Currency{
			Symbol:   tokenInfo.TokenSymbol,
			Decimals: int32(tokenInfo.Decimals),
			Metadata: map[string]interface{}{
				"tti": tokenId.Hex(),
			},
		}
		balances[i] = &types.Amount{
			Value:    value.String(),
			Currency: currency,
		}
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Hash:  block.Hash.Hex(),
			Index: int64(block.Height),
		},
		Balances: balances,
		// Metadata: map[string]interface{}{
		// 	"address":    address.Hex(),
		// 	"blockCount": 0,
		// },
	}, nil
}

// Send a transaction to the blockchain
func (ec *Client) SendTransaction(ctx context.Context, tx *api.AccountBlock) error {
	return ec.c.SendRawTransaction(ctx, tx)
}
