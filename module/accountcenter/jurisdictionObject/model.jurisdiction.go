package jurisdictionObject

func (Jurisdiction) TableName() string {
	return "jurisdictions"
}
func (*Jurisdiction) CacheExpire() int64 {
	return 5000
}

type Jurisdiction struct {
	ID   int64  `json:"id"`
	Fid  int64  `json:"fid"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func (*Jurisdictions) CacheExpire() int64 {
	return 5000
}

type Jurisdictions struct {
	ID     int64           `json:"id"`
	Fid    int64           `json:"fid"`
	Code   string          `json:"code"`
	Name   string          `json:"name"`
	Childs []Jurisdictions `json:"childs" xorm:"-" gorm:"-"`
}
