package towgo

import "errors"

func Call(method, token string, requestParams any, responseParams any) (err error) {
	request := NewJsonrpcrequest()
	request.Method = method
	request.Params = requestParams
	request.Session = token

	if token != "" {
		return CallGateWay(method, token, requestParams, responseParams)
	}

	//检查本地method
	if HasMethod(method) {
		localRpcConn := NewLocalRpcConnection(nil, nil)
		localRpcConn.Call(request, func(jrc JsonRpcConnection) {
			resp := jrc.GetRpcResponse()
			if resp != nil {
				if resp.Error.Code != 200 {
					err = errors.New(resp.Error.Message)
					return
				}
			}

			if responseParams != nil {
				err = jrc.ReadResult(responseParams)
			}

		})
		return
	}
	return CallGateWay(method, token, requestParams, responseParams)
}
