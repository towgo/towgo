package dblog

var NotSaveLogMethod = []string{
	"/account/myinfo",
	"/account/list",
	"/account/listbyidentity",
	"/account/identity/query",
	"/account/UnauthorizedMethodList",
	"/account/getAccountInfoByUsername",
	"/get/account/under/department",
	"/group/list",
	"/account/department/list",
	"/account/department/treelist",
	"/account/department/children",
	"/account/department/one/level/children/",
	"/account/identity/detail",
	"/account/identity/list",
	"/account/identity/list_by_jurisdiction_code",
	"/account/jurisdiction/detail",
	"/account/jurisdiction/list",
	"/account/jurisdiction/treelist",
	"/system/log/create",
	"/system/log/list",
	"/account/query",
	"/nav/tree",
}

const (
	CREATE    int64 = 1
	QUERY     int64 = 2
	UPDATE    int64 = 3
	DOWNLOAD  int64 = 4
	IMPORT    int64 = 5
	OTHER     int64 = 6
	DELETE    int64 = 7
	LOGIN     int64 = 8
	LOGIN_OUT int64 = 9
	UPLOAD    int64 = 10
	EXPORT    int64 = 11
)
const (
	ERROR   int64 = -1
	SUCCESS int64 = 0
)
