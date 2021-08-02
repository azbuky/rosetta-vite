package rpc

import (
	"context"

	"github.com/vitelabs/go-vite/rpc"
)

type UtilApi interface {
	GetPoWNonce(ctx context.Context, difficulty string, hash string) (string, error)
}

type utilApi struct {
	cc *rpc.Client
}

func NewUtilApi(cc *rpc.Client) UtilApi {
	return &utilApi{cc: cc}
}

func (ui utilApi) GetPoWNonce(
	ctx context.Context, 
	difficulty string, 
	hash string,
) (nonce string, err error) {
	err = ui.cc.CallContext(ctx, &nonce, "util_getPoWNonce", difficulty, hash)
	return
}
