/*
json rpc 拦截器
by:liangliangit
*/
package towgo

var handller []func(conn JsonRpcConnection) error

// 创建拦截器
func AddInterceptor(args ...func(conn JsonRpcConnection) error) {
	handller = append(handller, args...)
}

// 拦截器
// 拦截相关信息  如果返回的error 为空  说明没有拦截
func defaultJsonRpcInterceptor(conn JsonRpcConnection) error {
	for _, v := range handller {
		err := v(conn)
		if err != nil {
			return err
		}
	}
	return nil
}
