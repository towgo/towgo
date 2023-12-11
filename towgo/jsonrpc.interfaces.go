package towgo

// jsonrpc连接器接口(双向)
type JsonRpcConnection interface {

	//获取远程客户端IP
	GetRemoteAddr() string

	//读取原生的json字符串
	Read() string

	//将参数映射到传入的指针
	ReadParams(destParams ...interface{}) error

	//将结果映射到传入的指针
	ReadResult(destResult ...interface{}) error

	//将结果写入（最终会组装成一个响应对象发送给对端）
	WriteResult(any interface{})

	//写入连接器内置的响应对象
	Write()

	//直接将传入的响应对象写入
	WriteResponse(Jsonrpcresponse)

	//获取连接器请求对象
	GetRpcRequest() *Jsonrpcrequest

	//获取连接器响应对象
	GetRpcResponse() *Jsonrpcresponse

	/*
		推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
		push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
	*/
	Push(request *Jsonrpcrequest) error

	/*
		全双工模式下可以作为客户端进行请求通讯 注意 Call方法是异步的
	*/

	Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection))

	//连接器的底层协议类型 tcp|http|websocket
	LinkType() string

	IsConnected() bool

	GUID() string

	//开启心跳检测
	EnableHealthCheck()

	//关闭心跳检测
	DisableHealthCheck()

	WriteError(code int64, msg string)

	Close()
}
