package models

type TableQuery struct {
	TableName string `json:"tableName"`
	Query     Query  `json:"query"`
}
type Query struct {
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Field   []string            `json:"field"`
	Orderby []map[string]string `json:"orderby"`
	Filters []condition         `json:"filters"` //条件
}

type condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}
