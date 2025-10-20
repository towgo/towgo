/*
json rpc 拦截器
by:liangliangit
*/
package towgo

import "log"

var interceptorHandller []func(conn JsonRpcConnection) error

var passersHandller []func(conn JsonRpcConnection) error

// 创建拦截器
func AddInterceptor(args ...func(conn JsonRpcConnection) error) {
	interceptorHandller = append(interceptorHandller, args...)
}

// 创建放行器(满足放行条件的,优先放行,不会被拦截器拦截)
// 创建拦截器
func AddPassers(args ...func(conn JsonRpcConnection) error) {
	passersHandller = append(passersHandller, args...)
}

// 拦截器
// 拦截相关信息  如果返回的error 为空  说明没有拦截
func defaultJsonRpcInterceptor(conn JsonRpcConnection) error {

	//有放行器,优先执行
	for _, v := range passersHandller {
		err := v(conn)
		if err != nil {
			log.Print(err.Error())
			continue
		}
		log.Print("放行成功")
		return nil
	}

	for _, v := range interceptorHandller {
		err := v(conn)
		if err != nil {
			return err
		}
	}
	return nil
}
