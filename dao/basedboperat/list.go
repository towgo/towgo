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

func (lop *listOperatParams) SetPage(page int) {
	lop.Page = page
}
func (lop *listOperatParams) SetLimit(limit int) {
	lop.Limit = limit
}
func (lop *listOperatParams) SetField(fields []string) {
	if lop.Field == nil || len(fields) <= 0 {
		lop.Field = []string{}
	}
	lop.Field = append(lop.Field, fields...)
}
func (lop *listOperatParams) AddOrderby(field, orderby string) {
	if lop.Orderby == nil || len(lop.Orderby) <= 0 {
		lop.Orderby = []map[string]string{}
	}
	lop.Orderby = append(lop.Orderby, map[string]string{field: orderby})
}
func (lop *listOperatParams) AddJoin(field string, v []interface{}) {
	if lop.Join == nil || len(lop.Join) <= 0 {
		lop.Join = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Join[field] = append(lop.Join[field], v...)
}
func (lop *listOperatParams) AddAnd(field string, v []interface{}) {
	if lop.And == nil || len(lop.And) <= 0 {
		lop.And = map[string][]interface{}{field: []interface{}{}}
	}
	lop.And[field] = v
}
func (lop *listOperatParams) AddOr(field string, v []interface{}) {
	if lop.Or == nil || len(lop.Or) <= 0 {
		lop.Or = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Or[field] = append(lop.Or[field], v...)
}
func (lop *listOperatParams) AddNot(field string, v []interface{}) {
	if lop.Not == nil || len(lop.Not) <= 0 {
		lop.Not = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Not[field] = append(lop.Not[field], v...)
}
func (lop *listOperatParams) AddLike(field string, v []interface{}) {
	if lop.Like == nil || len(lop.Like) <= 0 {
		lop.Like = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Like[field] = append(lop.Like[field], v...)

}
func (lop *listOperatParams) AddAndLike(field string, v []interface{}) {
	if lop.AndLike == nil || len(lop.AndLike) <= 0 {
		lop.AndLike = map[string][]interface{}{field: []interface{}{}}
	}
	lop.AndLike[field] = append(lop.AndLike[field], v...)
}
func (lop *listOperatParams) AddOrLike(field string, v []interface{}) {
	if lop.OrLike == nil || len(lop.OrLike) <= 0 {
		lop.OrLike = map[string][]interface{}{field: []interface{}{}}
	}
	lop.OrLike[field] = append(lop.OrLike[field], v...)
}
func (lop *listOperatParams) AddWhere(field, operator string, value interface{}) {
	if lop.Where == nil {
		lop.Where = []Condition{}
	}
	lop.Where = append(lop.Where, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})

}
