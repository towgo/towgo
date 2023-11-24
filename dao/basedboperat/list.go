package basedboperat

type List struct {
	CacheExpireLimit int64
	Count            int64 `json:"count"`
	Error            error `json:"-"`
	listOperatParams
}
type listOperatParams struct {
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Field   []string            `json:"field"`
	Orderby []map[string]string `json:"orderby"`
	Join    map[string][]interface{}
	And     map[string][]interface{} `json:"and"`
	Or      map[string][]interface{} `json:"or"`
	Not     map[string][]interface{} `json:"not"`
	Like    map[string][]interface{} `json:"like"`
	AndLike map[string][]interface{} `json:"andlike"`
	OrLike  map[string][]interface{} `json:"orlike"`
	Where   []Condition              `json:"where"` //条件
}

type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type ListSimple struct {
	In      map[string][]interface{} `json:"and"`
	Table   string                   `json:"table"`
	Count   int64                    `json:"count"`
	Error   error                    `json:"-"`
	Field   []string                 `json:"field"`
	Orderby []map[string]string      `json:"orderby"`
}
