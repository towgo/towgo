package accountcenter

func (Userorderinfo) TableName() string {
	return "userorderinfo"
}

type Userorderinfo struct {
	ID          int64  `json:"id"`
	Uiid        int64  `json:"uiid"`
	Shoujianren string `json:"shoujianren"`
	Mobile      string `json:"mobile"`
}
