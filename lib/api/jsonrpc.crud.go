/*
code by liangliangit

通过chatGPT辅助
实现了模型与增删改查列表的分离
可以为模型增加增删改查接口
*/
package api

import (
	"context"
	"reflect"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
)

type MyKey string

const (
	CRUD_FLAG_CREATE = "CRUD_FLAG_CREATE"
	CRUD_FLAG_DELETE = "CRUD_FLAG_DELETE"
	CRUD_FLAG_UPDATE = "CRUD_FLAG_UPDATE"
	CRUD_FLAG_DETAIL = "CRUD_FLAG_DETAIL"
	CRUD_FLAG_LIST   = "CRUD_FLAG_LIST"
)

type Crud struct {
	baseMethod   string
	modelObject  interface{}
	modelObjects interface{}
}

func NewCRUDJsonrpcAPI(baseMethod string, modelObject, modelObjects interface{}) *Crud {
	return &Crud{
		baseMethod:   baseMethod,
		modelObject:  modelObject,
		modelObjects: modelObjects,
	}
}

func (c *Crud) create(rpcConn jsonrpc.JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	rpcConn.ReadParams(&model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey jsonrpc.ContextKey = jsonrpc.JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	_, err = session.Create(model)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func (c *Crud) update(rpcConn jsonrpc.JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	m := map[string]interface{}{}
	rpcConn.ReadParams(&model, &m)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey jsonrpc.ContextKey = jsonrpc.JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	err = session.Update(model, m, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func (c *Crud) delete(rpcConn jsonrpc.JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	rpcConn.ReadParams(&model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey jsonrpc.ContextKey = jsonrpc.JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	count, err := session.Delete(model, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(count)
}

func (c *Crud) detail(rpcConn jsonrpc.JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	rpcConn.ReadParams(&model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey jsonrpc.ContextKey = jsonrpc.JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	err = session.Get(model, nil, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func (c *Crud) list(rpcConn jsonrpc.JsonRpcConnection) {
	modelType := c.modelObject
	modelTypes := c.modelObjects
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	models := reflect.New(reflect.TypeOf(modelTypes)).Interface()

	var list basedboperat.List
	rpcConn.ReadParams(&list)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey jsonrpc.ContextKey = jsonrpc.JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	session.ListScan(&list, model, models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

/*
reg jsonrpc method
if CRUD_FLAG is null ,Will Reg All Method
*/
func (c *Crud) RegAPI(CRUD_FLAG ...string) {
	if CRUD_FLAG == nil {
		c.regAPI(CRUD_FLAG_CREATE, CRUD_FLAG_DELETE, CRUD_FLAG_DETAIL, CRUD_FLAG_LIST, CRUD_FLAG_UPDATE)
		return
	}
	c.regAPI(CRUD_FLAG...)
}

func (c *Crud) regAPI(CRUD_FLAG ...string) {
	for _, v := range CRUD_FLAG {
		switch v {
		case CRUD_FLAG_CREATE:
			jsonrpc.SetFunc(c.baseMethod+"/create", c.create)
		case CRUD_FLAG_DELETE:
			jsonrpc.SetFunc(c.baseMethod+"/delete", c.delete)
		case CRUD_FLAG_DETAIL:
			jsonrpc.SetFunc(c.baseMethod+"/detail", c.detail)
		case CRUD_FLAG_LIST:
			jsonrpc.SetFunc(c.baseMethod+"/list", c.list)
		case CRUD_FLAG_UPDATE:
			jsonrpc.SetFunc(c.baseMethod+"/update", c.update)
		}
	}
}
