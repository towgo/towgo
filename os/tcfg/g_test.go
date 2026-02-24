package tcfg

import (
	"github.com/towgo/towgo/lib/utils"
	"log"
	"testing"
)

// TestNewAdapter 测试 Adapter 创建功能
func TestNewAdapter(t *testing.T) {

	adapter, err := NewAdapter()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(adapter.GetConfigPath())
	err = adapter.LoadConfig()
	if err != nil {
		t.Fatal(err)
		return
	}
	log.Println(adapter.Data())
	key, value := utils.MapPossibleItemByKey(adapter.Data(), "server")
	log.Println(key, value)
	//	 map 转 结构体
	var s Server
	utils.MapToStruct(
		value,
		&s,
	)
	log.Println(s)
	get, err := adapter.Get("database")
	if err != nil {
		t.Fatal(err)
		return
	}
	log.Println(get)
}

type T struct {
	DbType          string `json:"DbType"`
	IsMaster        bool   `json:"IsMaster"`
	Dsn             string `json:"Dsn"`
	SqlMaxIdleConns int    `json:"sqlMaxIdleConns"`
	SqlMaxOpenConns int    `json:"sqlMaxOpenConns"`
	SqlLogLevel     int    `json:"sqlLogLevel"`
}
type Server struct {
	ServerportTls string `json:"serverport_tls"`
	TlsKeyPath    string `json:"tls_key_path"`
	TlsCertPath   string `json:"tls_cert_path"`
	Serverport    string `json:"serverport"`
	ManagerPort   string `json:"managerPort"`
	MailInform    bool   `json:"mail_inform"`
}
