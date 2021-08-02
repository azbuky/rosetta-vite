package rpc

import (
	"context"

	"github.com/vitelabs/go-vite/net"
	"github.com/vitelabs/go-vite/rpc"
	"github.com/vitelabs/go-vite/rpcapi/api"
)

type NetApi interface {
	GetNodeInfo(ctx context.Context) (*net.NodeInfo, error)
	GetSyncInfo(ctx context.Context) (*api.SyncInfo, error)
}

type netApi struct {
	cc *rpc.Client
}

func NewNetApi(cc *rpc.Client) NetApi {
	return &netApi{cc: cc}
}

func (ni netApi) GetSyncInfo(ctx context.Context) (syncInfo *api.SyncInfo, err error) {
	syncInfo = &api.SyncInfo{}
	err = ni.cc.CallContext(ctx, syncInfo, "net_syncInfo")
	return
}

func (ni netApi) GetNodeInfo(ctx context.Context) (nodeInfo *net.NodeInfo, err error) {
	nodeInfo = &net.NodeInfo{} 
	err = ni.cc.CallContext(ctx, nodeInfo, "net_nodeInfo")
	return
}