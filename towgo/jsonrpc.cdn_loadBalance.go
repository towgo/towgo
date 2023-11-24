package towgo

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const (
	LOAD_ALGORITHM_RAND     = 0 //随机算法
	LOAD_ALGORITHM_PRIORITY = 1 //优先级算法
	LOAD_ALGORITHM_POLLING  = 2 //轮询算法
	LOAD_ALGORITHM_WEIGHT   = 3 //权重算法
)

/*
网关负载均衡
by:liangliangit
*/
type LoadBalance struct {
	sync.Mutex
	methodWebsocketJsonrpcRandConnMap     sync.Map //服务与websocket jsonrpc连接随机路由关系映射表
	methodWebsocketJsonrpcConnMapPriority sync.Map //服务与websocket jsonrpc连接优先级路由关系映射表
	methodHttpRestfulRandConnMap          sync.Map //服务与http restful 连接随机路由关系映射表
}

type EdgeServerNodeHttpRoute struct {
	Guid    string `json:"guid"`
	HostUrl string `json:"host_url"` // eg:http://192.168.1.10:8090
}

type LoadBalancePrioritys []*LoadBalancePriority

type LoadBalancePriority struct {
	Priority int64
	RpcConn  JsonRpcConnection
}

func (m LoadBalancePrioritys) Len() int {
	return len(m)
}

// 生序 （从小到大）
func (m LoadBalancePrioritys) Less(x, y int) bool {
	return m[x].Priority < m[y].Priority
}

// swap 进行位置置换
func (m LoadBalancePrioritys) Swap(x, y int) {
	m[x], m[y] = m[y], m[x]
}
func switch_config(rpcConn JsonRpcConnection, e EdgeServerNodeConfig) *EdgeServerNodeHttpRoute {
	var es EdgeServerNodeHttpRoute
	if e.Port == "" {
		e.Port = "80"
	}
	if e.Schema == "" {
		e.Schema = "http"
	}

	if e.EdgeServerNodeHost == "" {
		hostarr := strings.Split(rpcConn.GetRemoteAddr(), ":")
		e.EdgeServerNodeHost = hostarr[0]
	}

	es.Guid = rpcConn.GUID()
	es.HostUrl = e.Schema + "://" + e.EdgeServerNodeHost + ":" + e.Port
	return &es
}

// 新增http restful 短连接路由
func (lb *LoadBalance) StoreHttpRestfulMethod(method string, rpcConn JsonRpcConnection, e EdgeServerNodeConfig) {
	lb.Lock()
	defer lb.Unlock()
	//增加随机路由
	connMapInterface, ok := lb.methodHttpRestfulRandConnMap.Load(method)
	var connMap *sync.Map
	if ok {
		connMap = connMapInterface.(*sync.Map)
	} else {
		connMap = &sync.Map{}
	}

	connMap.Store(rpcConn.GUID(), switch_config(rpcConn, e))
	lb.methodHttpRestfulRandConnMap.Store(method, connMap)
}

// 新增http jsonrpc 短连接路由
func (lb *LoadBalance) StoreHttpJsonrpcMethod(method string, rpcConn JsonRpcConnection, e EdgeServerNodeConfig) {

}

// 新增jsonrpc 长连接路由 (tcp 或 websocket)
func (lb *LoadBalance) StoreJsonrpcKeepaliveMethod(method string, rpcConn JsonRpcConnection, priority int64) {
	lb.Lock()
	defer lb.Unlock()

	//增加随机路由
	connMapInterface, ok := lb.methodWebsocketJsonrpcRandConnMap.Load(method)
	var connMap *sync.Map
	if ok {
		connMap = connMapInterface.(*sync.Map)
	} else {
		connMap = &sync.Map{}
	}
	connMap.Store(rpcConn.GUID(), rpcConn)
	lb.methodWebsocketJsonrpcRandConnMap.Store(method, connMap)

	//增加优先级路由
	connMapPriorityInterface, ok := lb.methodWebsocketJsonrpcConnMapPriority.Load(method)
	var connMapPriority LoadBalancePrioritys
	l := &LoadBalancePriority{}
	l.Priority = priority
	l.RpcConn = rpcConn
	if ok {
		connMapPriority = connMapPriorityInterface.(LoadBalancePrioritys)
		connMapPriority = append(connMapPriority, l)
		sort.Sort(connMapPriority)
	} else {
		connMapPriority = append(connMapPriority, l)
	}
	lb.methodWebsocketJsonrpcConnMapPriority.Store(method, connMapPriority)

}

func (lb *LoadBalance) LoadPriority(method string) JsonRpcConnection {
	connMapPriorityInterface, ok := lb.methodWebsocketJsonrpcConnMapPriority.Load(method)
	if ok {
		connMapPriority := connMapPriorityInterface.(LoadBalancePrioritys)
		if len(connMapPriority) > 0 {
			return connMapPriority[0].RpcConn
		}
	}
	return nil
}

func (lb *LoadBalance) RandPickRestfulProxy(r *http.Request) *EdgeServerNodeHttpRoute {

	log.Print(r.URL.Path)
	es := lb.LoadRandHttpRestful(r.URL.Path)

	//完全匹配未命中
	if es == nil {
		//模糊匹配
		paths := strings.Split(r.URL.Path, "/")
		var mathPaths []string
		var path string
		pathsLen := len(paths)

		mathPaths = append(mathPaths, "/")

		for i := 0; i < pathsLen; i++ {
			if paths[i] != "" {
				path = path + "/" + paths[i]
			}

			mathPaths = append(mathPaths, path)
		}

		for i := len(mathPaths) - 1; i >= 0; i-- {
			es = lb.LoadRandHttpRestful(mathPaths[i])
			if es != nil {
				break
			}
		}
	}
	return es

}

func (lb *LoadBalance) LoadRandHttpRestful(method string) *EdgeServerNodeHttpRoute {
	connMapInterface, ok := lb.methodHttpRestfulRandConnMap.Load(method)
	if ok {
		connMap := connMapInterface.(*sync.Map)
		var es *EdgeServerNodeHttpRoute
		connMap.Range(func(key, value any) bool {
			es = value.(*EdgeServerNodeHttpRoute)
			return false
		})
		return es
	}
	return nil
}

func (lb *LoadBalance) LoadRand(method string) JsonRpcConnection {
	connMapInterface, ok := lb.methodWebsocketJsonrpcRandConnMap.Load(method)
	if ok {
		connMap := connMapInterface.(*sync.Map)
		var rpcConn JsonRpcConnection
		connMap.Range(func(key, value any) bool {
			rpcConn = value.(JsonRpcConnection)
			return false
		})
		return rpcConn
	}
	return nil
}

func (lb *LoadBalance) LoadPolling(method string) JsonRpcConnection {
	return nil
}

func (lb *LoadBalance) RemoveByGuid(rpcConnGuid string) {
	lb.Lock()
	defer lb.Unlock()
	lb.methodWebsocketJsonrpcRandConnMap.Range(func(key, value any) bool {
		connMap := value.(*sync.Map)
		if connMap != nil {
			connMap.Delete(rpcConnGuid)
		}
		return true
	})

	lb.methodHttpRestfulRandConnMap.Range(func(key, value any) bool {
		connMap := value.(*sync.Map)
		if connMap != nil {
			connMap.Delete(rpcConnGuid)
		}
		return true
	})

	lb.methodWebsocketJsonrpcConnMapPriority.Range(func(key, value any) bool {
		connMapPriority := value.(LoadBalancePrioritys)
		newMapPriority := LoadBalancePrioritys{}
		for _, v := range connMapPriority {
			if v.RpcConn.GUID() != rpcConnGuid {
				newMapPriority = append(newMapPriority, v)
			}
		}
		sort.Sort(newMapPriority)
		lb.methodWebsocketJsonrpcConnMapPriority.Store(key.(string), newMapPriority)
		return true
	})
}

func (lb *LoadBalance) RemoveMethod(method string) {
	lb.Lock()
	defer lb.Unlock()
	lb.methodWebsocketJsonrpcRandConnMap.Delete(method)
	lb.methodWebsocketJsonrpcConnMapPriority.Delete(method)
}
