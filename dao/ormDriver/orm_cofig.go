package ormDriver

var ConfigNodeNameDatabase = "databases"

type Config struct {
	Mode  string   `json:"mode"`
	Nodes []DbNode `json:"nodes"`
}
type DbNode struct {
	DbType          string
	IsMaster        bool
	Dsn             string
	SqlLogLevel     int64
	SqlMaxIdleConns int64
	SqlMaxOpenConns int64
}
