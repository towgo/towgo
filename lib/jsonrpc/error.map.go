package jsonrpc

type codeTypes map[int64]string

var CODETYPES = NewCodeTypes()

func NewCodeTypes() codeTypes {
	c := codeTypes{
		0:    "未知错误",
		1001: "账户名不能为空",
		1002: "密码不能为空",
		1003: "账户未授权",
		1004: "账户不存在",
		1005: "密码错误",
		1006: "账户已经存在",
		2000: "数据库连接失败",
		2001: "未查询到相关记录",

		-32700: "Parse error语法解析错误,服务端接收到无效的json。该错误发送于服务器尝试解析json文本",
		-32600: "Invalid Request无效请求 请求json 的版本不正确,json-rpc版本为2.0",
		-32601: "Method not found找不到方法",
		-32000: "Server error服务端错误",
	}
	return c
}
