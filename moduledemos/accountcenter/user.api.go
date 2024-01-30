package accountcenter

import (
	"log"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/towgo"
)

func init() {
	xormDriver.Sync2(new(User), new(Userinfo), new(Userorderinfo))
	towgo.NewCRUDJsonrpcAPI("/user", User{}, []User{}).RegAPI()
	towgo.SetFunc("/user/login", login)
	towgo.SetFunc("/user/list2", list)
	towgo.SetFunc("/user/get", get)
}

func update(rpcConn towgo.JsonRpcConnection) {
	var u User
	m := map[string]interface{}{}

	rpcConn.ReadParams(&u, &m)
	basedboperat.Update(&u, &m, "id = ?", u.ID)
	rpcConn.WriteResult("ok")

}

func get(rpcConn towgo.JsonRpcConnection) {
	var u User
	//var us []User

	var umap map[string]interface{} = map[string]interface{}{}

	basedboperat.SqlQueryScan(&umap, "select * from "+u.TableName()+" where id = ?", 1)
	rpcConn.WriteResult(umap)
}

func login(rpcConn towgo.JsonRpcConnection) {
	var user, findUser User
	rpcConn.ReadParams(&user)

	//通过用户名查询用户数据
	err := basedboperat.Get(&findUser, nil, "username = ?", user.Username)

	if err != nil {
		//不一致:返回错误
		rpcConn.WriteError(500, err.Error())
		return
	}
	//检查用户名是否存在

	if findUser.ID == 0 {
		//不一致:返回错误
		rpcConn.WriteError(500, "用户名不存在")
		return
	}

	//加密密码
	upassword := system.MD5(user.Password)

	//撒盐
	upassword = upassword + findUser.Salt

	//混合加密
	upassword = system.MD5(upassword)

	//判断密码是否一致
	if findUser.Password != upassword {
		//不一致:返回错误
		rpcConn.WriteError(500, "密码错误")
		return
	}

	token := system.MD5(system.GetGUID().Hex())

	rpcConn.WriteResult(token)

}

func list(rpcConn towgo.JsonRpcConnection) {

	var params struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		ID    int64  `json:"id"`
	}

	rpcConn.ReadParams(&params)

	log.Print(params)
	rpcConn.WriteResult(params)

	return
	var users []User

	var user *User
	log.Print(user)

	basedboperat.SqlQueryScan(&users, "select * from user where id = 0")

	if len(users) == 0 {
		log.Print("空")
	} else {
		log.Print("有内容")
	}

	rpcConn.WriteResult(users)

}
