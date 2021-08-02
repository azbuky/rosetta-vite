package rpc

import (
	"github.com/vitelabs/go-vite/rpc"
)

type RpcClient interface {
	LedgerApi
	ContractApi
	NetApi
	UtilApi

	GetClient() *rpc.Client
}

func NewRpcClient(rawurl string) (RpcClient, error) {
	c, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}

	r := &rpcClient{
		LedgerApi:   NewLedgerApi(c),
		ContractApi: NewContractApi(c),
		NetApi: 	 NewNetApi(c),
		UtilApi: 	 NewUtilApi(c),
		cc:          c,
	}
	return r, nil
}

type rpcClient struct {
	LedgerApi
	ContractApi
	NetApi
	UtilApi

	cc *rpc.Client
}

func (c rpcClient) GetClient() *rpc.Client {
	return c.cc
}
