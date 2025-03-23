package basedboperat

type DbContext interface {
	WithValue(key, value any)
	Value(key any) any
	HasValue(key, value any) bool
}

// 通用数据库基本操作接口
type DbOperat interface {
	DbContext
	//获取第一条记录
	First(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error

	//获取最后一条记录
	Last(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error

	//列表查询
	ListScan(l *List, model interface{}, destModels interface{})

	//获取一条记录
	Get(destModel interface{}, selectFields []string, condition interface{}, conditionArgs ...interface{}) error

	//修改记录
	Update(model interface{}, fields any, condition interface{}, conditionArgs ...interface{}) error

	//删除记录
	Delete(model interface{}, PrimaryKeyID interface{}, condition interface{}, conditionArgs ...interface{}) (int64, error)

	//创建记录
	Create(model interface{}) (int64, error)

	//执行原生sql命令
	SqlExec(sql interface{}, args ...interface{}) error

	//执行原生sql查询
	SqlQuery(sql interface{}, args ...interface{}) (resultsSlice []map[string]interface{}, err error)

	//执行原生sql查询并将结果解析到结构体、map等结构中
	SqlQueryScan(destModel interface{}, sql interface{}, args ...interface{}) error

	//执行根据条件查询
	QueryScan(destModel interface{}, order interface{}, condition interface{}, args ...interface{}) error

	//获取记录数
	Count(model interface{}, intPrt *int64, condition interface{}, conditionArgs ...interface{}) error

	//获取当前选择的字段
	GetCurrentSelectFields() []string

	//当前是否选中该字段
	IsCurrentSelectedField(field string) bool

	//同步结构到数据库
	Sync(beans ...any)
}

// 事务接口
type DbTransactionSession interface {
	Begin() error
	Commit() error
	Rollback() error
	DbOperat
}
