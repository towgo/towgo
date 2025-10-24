package model

// PageReq 公共请求参数
type PageReq struct {
	DateRange []string `p:"dateRange"`       //日期范围
	PageNum   int      `p:"pageNum" d:"1"`   //当前页码
	PageSize  int      `p:"pageSize" d:"10"` //每页数
	OrderBy   string   //排序方式
}

// ListRes 列表公共返回
type ListRes struct {
	PageNum  int         `json:"pageNum"`
	PageSize int         `json:"pageSize"`
	Count    int         `json:"count"`
	Rows     interface{} `json:"rows"`
}
