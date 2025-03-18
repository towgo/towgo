package accountcenter

func (DepartmentJurisdiction) TableName() string {
	return "departments_jurisdictions"
}
func (*DepartmentJurisdiction) CacheExpire() int64 {
	return 5000
}

type DepartmentJurisdiction struct {
	ID             int64
	DepartmentId   int64
	JurisdictionId int64
}
