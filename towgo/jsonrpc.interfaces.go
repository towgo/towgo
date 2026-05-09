package towgo

import "context"

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

	//设置结果（不发送，由中间件决定何时 Write）
	SetResult(result interface{})

	//获取已设置的结果
	GetResult() interface{}

	//设置结果并立即发送响应
	WriteResult(result interface{})

	//写入响应（发送给对端）
	Write()

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

	//设置错误并发送错误响应
	WriteError(code int64, msg string)

	//直接写入响应对象
	WriteResponse(resp Jsonrpcresponse)

	Close()

	SetValue(key string, value any)
	GetValue(key string) (value any, ok bool)

	// Context  获取上下文
	Context() context.Context

	// WithContext 设置上下文（用于中间件传递修改后的 ctx）
	WithContext(ctx context.Context)

	// GetError 获取错误
	GetError() error

	// WithError 设置错误
	WithError(err error)

	// Next 继续执行下一个中间件/handler
	Next()
}
