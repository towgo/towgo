package basedboperat

type List struct {
	CacheExpireLimit int64
	Count            int64 `json:"count"`
	Error            error `json:"-"`
	ListOperatParams
}
type ListOperatParams struct {
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

func (lop *ListOperatParams) SetPage(page int) *ListOperatParams {
	lop.Page = page
	return lop
}
func (lop *ListOperatParams) SetLimit(limit int) *ListOperatParams {
	lop.Limit = limit
	return lop
}
func (lop *ListOperatParams) SetField(fields []string) *ListOperatParams {
	if lop.Field == nil || len(fields) <= 0 {
		lop.Field = []string{}
	}
	lop.Field = append(lop.Field, fields...)
	return lop
}
func (lop *ListOperatParams) AddOrderby(field, orderby string) *ListOperatParams {
	if lop.Orderby == nil || len(lop.Orderby) <= 0 {
		lop.Orderby = []map[string]string{}
	}
	lop.Orderby = append(lop.Orderby, map[string]string{field: orderby})
	return lop
}
func (lop *ListOperatParams) AddJoin(field string, v []interface{}) *ListOperatParams {
	if lop.Join == nil || len(lop.Join) <= 0 {
		lop.Join = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Join[field] = append(lop.Join[field], v...)
	return lop
}
func (lop *ListOperatParams) AddAnd(field string, v []interface{}) *ListOperatParams {
	if lop.And == nil || len(lop.And) <= 0 {
		lop.And = map[string][]interface{}{field: []interface{}{}}
	}
	lop.And[field] = v
	return lop
}
func (lop *ListOperatParams) AddOr(field string, v []interface{}) *ListOperatParams {
	if lop.Or == nil || len(lop.Or) <= 0 {
		lop.Or = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Or[field] = append(lop.Or[field], v...)
	return lop
}
func (lop *ListOperatParams) AddNot(field string, v []interface{}) *ListOperatParams {
	if lop.Not == nil || len(lop.Not) <= 0 {
		lop.Not = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Not[field] = append(lop.Not[field], v...)
	return lop
}
func (lop *ListOperatParams) AddLike(field string, v []interface{}) *ListOperatParams {
	if lop.Like == nil || len(lop.Like) <= 0 {
		lop.Like = map[string][]interface{}{field: []interface{}{}}
	}
	lop.Like[field] = append(lop.Like[field], v...)
	return lop

}
func (lop *ListOperatParams) AddAndLike(field string, v []interface{}) *ListOperatParams {
	if lop.AndLike == nil || len(lop.AndLike) <= 0 {
		lop.AndLike = map[string][]interface{}{field: []interface{}{}}
	}
	lop.AndLike[field] = append(lop.AndLike[field], v...)
	return lop
}
func (lop *ListOperatParams) AddOrLike(field string, v []interface{}) *ListOperatParams {
	if lop.OrLike == nil || len(lop.OrLike) <= 0 {
		lop.OrLike = map[string][]interface{}{field: []interface{}{}}
	}
	lop.OrLike[field] = append(lop.OrLike[field], v...)
	return lop
}
func (lop *ListOperatParams) AddWhere(field, operator string, value interface{}) *ListOperatParams {
	if lop.Where == nil {
		lop.Where = []Condition{}
	}
	lop.Where = append(lop.Where, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})

	return lop
}
