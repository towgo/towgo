package accountcenter

import "errors"

func (User) TableName() string {
	return "user"
}

func (u *User) InputCheck() error {
	if u.Username == "" {
		return errors.New("用户名不能为空")
	}

	if u.Username == "abc" {
		return errors.New("用户名不能abc")
	}

	return nil
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
