package accountcenter

func (User) TableName() string {
	return "user"
}

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Salt      string     `json:"-"`
	Token     string     `json:"token" xorm:"-" gorm:"-"`
	UserToken *UserToken `json:"-" gorm:"-" xorm:"-"`
	Userinfo  Userinfo   `json:"userinfo" xorm:"-"`
}
