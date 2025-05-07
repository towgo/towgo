package tcfg

import "github.com/towgo/towgo/os/log"

var conf *Config

func init() {
	config, err := New()
	if err != nil {
		log.Error(err)
		return
	}
	conf = config
	if err = conf.LoadConfig(); err != nil {
		log.Error(err)
		return
	}
}
func GetConfig() *Config {
	return conf
}

type Config struct {
	adapter *Adapter
}

func New() (*Config, error) {
	adapterFile, err := NewAdapter()
	if err != nil {
		return nil, err
	}
	return &Config{
		adapter: adapterFile,
	}, nil
}
func NewWithAdapter(adapter *Adapter) *Config {
	return &Config{
		adapter: adapter,
	}
}
func (c *Config) SetAdapter(adapter *Adapter) {
	c.adapter = adapter
}

func (c *Config) GetAdapter() *Adapter {
	return c.adapter
}

func (c *Config) SetConfigPath(path string) error {

	return c.adapter.SetConfigPath(path)
}
func (c *Config) GetConfigPath() string {
	return c.adapter.GetConfigPath()
}

// LoadConfig 读取并解析配置文件到嵌套的map结构
func (c *Config) LoadConfig() error {

	return c.adapter.LoadConfig()
}

// Data retrieves and returns all configuration data in current resource as map.
// Note that this function may lead lots of memory usage if configuration data is too large,
// you can implement this function if necessary.
func (c *Config) Data() map[string]interface{} {
	return c.adapter.Data()
}

// Get retrieves and returns value by specified `pattern` in current resource.
// Pattern like:
// "x.y.z" for map item.
// "x.0.y" for slice item.
func (c *Config) Get(pattern string) (interface{}, error) {
	return c.adapter.Get(pattern)
}
func (c *Config) GetDataToStruct(pattern string, p interface{}) (err error) {
	return c.adapter.GetDataToStruct(pattern, p)

}
