package towgo

type API struct {
	InputParams   any //入参
	OutputParams  any //出参
	RpcConnection JsonRpcConnection
	execFunc      func(api *API) error
}

func NewAPI() *API {
	var api API
	return &api
}
