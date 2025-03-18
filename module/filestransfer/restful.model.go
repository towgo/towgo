package filestransfer

type RestFulResult struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
	Status  int64       `json:"status"`
}
