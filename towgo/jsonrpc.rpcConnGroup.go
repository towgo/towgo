package towgo

import (
	"sync"

	"github.com/towgo/towgo/lib/system"
)

var defaultGroup *Group = &Group{}

var groups sync.Map

type Group struct {
	id       int64
	locker   sync.Mutex
	name     string
	ID       string
	rpcConns []JsonRpcConnection
}

func (g *Group) Join(rpcConn JsonRpcConnection) {
	g.locker.Lock()
	g.rpcConns = append(g.rpcConns, rpcConn)
	g.locker.Unlock()
}

func (g *Group) Leave(rpcConn JsonRpcConnection) {
	g.locker.Lock()
	g.rpcConns = append(g.rpcConns, rpcConn)
	g.locker.Unlock()
}

func (g *Group) PushToAll(r *Jsonrpcrequest) {
	for _, v := range g.rpcConns {
		v.Push(r)
	}
}

func NewGroup(groupName string) *Group {

	//检查名称是否重复
	g, ok := groups.Load(groupName)
	if ok {
		return g.(*Group)
	}

	group := &Group{
		ID:   system.GetGUID().Hex(),
		name: groupName,
	}
	return group
}

func PushToGroup(groupName string, request *Jsonrpcrequest) {

}

func PushToAll(r *Jsonrpcrequest) {
	defaultGroup.PushToAll(r)
}
