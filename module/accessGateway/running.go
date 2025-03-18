package accessGateway

/*
统一接入网关 运行时
by:liangliangit
*/
import "time"

func init() {
	go healthyCheck()
}

// 自动同步数据库数据
func autoSync() {
	for {
		time.Sleep(time.Second * 1)
		ReLoad()
	}
}

// 健康检查
func healthyCheck() {
	for {
		time.Sleep(time.Second * 1)
		for _, v := range tempApiRouteGroups {
			for _, way := range v.APIRouteWay {
				go func(way *ApiRouteWay) {
					way.HealthyCheck()
				}(way)
			}
		}
	}
}
