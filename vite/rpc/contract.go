package rpc

import (
	"context"

	"github.com/vitelabs/go-vite/rpc"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

type ContractApi interface {
	GetTokenInfoById(ctx context.Context, tokenId string) (*api.RpcTokenInfo, error)
}

type contractApi struct {
	cc *rpc.Client
}

func NewContractApi(cc *rpc.Client) ContractApi {
	return &contractApi{cc: cc}
}

func (ci contractApi) GetTokenInfoById(ctx context.Context, tokenId string) (tokenInfo *api.RpcTokenInfo, err error) {
	tokenInfo = &api.RpcTokenInfo{}
	err = ci.cc.CallContext(ctx, tokenInfo, "contract_getTokenInfoById", tokenId)
	return
}
