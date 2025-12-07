/*
code by liangliangit

通过chatGPT辅助
实现了模型与增删改查列表的分离
可以为模型增加增删改查接口
*/
package towgo

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/towgo/towgo/dao/basedboperat"
	"unicode"
)

const (
	CRUD_FLAG_CREATE = "CRUD_FLAG_CREATE"
	CRUD_FLAG_DELETE = "CRUD_FLAG_DELETE"
	CRUD_FLAG_UPDATE = "CRUD_FLAG_UPDATE"
	CRUD_FLAG_DETAIL = "CRUD_FLAG_DETAIL"
	CRUD_FLAG_LIST   = "CRUD_FLAG_LIST"
	IS_CRUD          = "IS_CRUD"
)

var crudMap map[string]*Crud = map[string]*Crud{}

type Crud struct {
	baseMethod           string
	CreateApi            *Api
	DeleteApi            *Api
	UpdateApi            *Api
	DetailApi            *Api
	ListApi              *Api
	CreateConditionFixed []func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)
	UpdateConditionFixed []func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)
	DeleteConditionFixed []func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)
	DetailConditionFixed []func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)
	ListConditionFixed   []func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)
	modelObject          interface{}
	modelObjects         interface{}
}

func NewCRUDJsonrpcAPI(baseMethod string, modelObject, modelObjects interface{}) *Crud {
	crud := &Crud{
		baseMethod:   baseMethod,
		modelObject:  modelObject,
		modelObjects: modelObjects,
	}

	crudMap[fmt.Sprint(modelObject)] = crud
	return crud
}

func RegListConditionFixed(modelObject interface{}, f func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)) {

	crud, ok := crudMap[fmt.Sprint(modelObject)]
	if !ok {
		return
	}
	crud.ListConditionFixed = append(crud.ListConditionFixed, f)
}

func RegDetailConditionFixed(modelObject interface{}, f func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)) {

	crud, ok := crudMap[fmt.Sprint(modelObject)]
	if !ok {
		return
	}
	crud.DetailConditionFixed = append(crud.DetailConditionFixed, f)
}

func RegDeleteConditionFixed(modelObject interface{}, f func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)) {

	crud, ok := crudMap[fmt.Sprint(modelObject)]
	if !ok {
		return
	}
	crud.DeleteConditionFixed = append(crud.DeleteConditionFixed, f)
}

func RegUpdateConditionFixed(modelObject interface{}, f func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)) {

	crud, ok := crudMap[fmt.Sprint(modelObject)]
	if !ok {
		return
	}
	crud.UpdateConditionFixed = append(crud.UpdateConditionFixed, f)
}
func RegCreateConditionFixed(modelObject interface{}, f func(rpcConn JsonRpcConnection) ([]basedboperat.Condition, error)) {

	crud, ok := crudMap[fmt.Sprint(modelObject)]
	if !ok {
		return
	}
	crud.CreateConditionFixed = append(crud.CreateConditionFixed, f)
}

func (c *Crud) create(rpcConn JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	rpcConn.ReadParams(&model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey ContextKey = JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	var crudKey ContextKey = IS_CRUD
	ctx = context.WithValue(ctx, crudKey, rpcConn)

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

func (c *Crud) update(rpcConn JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	m := map[string]interface{}{}
	rpcConn.ReadParams(&model, &m)
	var updateFildArray []string
	for k, _ := range m {
		var buf bytes.Buffer

		for i, r := range k {
			if unicode.IsUpper(r) && i > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		}
		updateFildArray = append(updateFildArray, buf.String())
	}
	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()

	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey ContextKey = JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	var crudKey ContextKey = IS_CRUD
	ctx = context.WithValue(ctx, crudKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	err = session.Update(model, updateFildArray, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func (c *Crud) delete(rpcConn JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()

	var groupInt []int64
	var groupString []string

	rpcConn.ReadParams(&model, &groupInt, &groupString)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey ContextKey = JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	var crudKey ContextKey = IS_CRUD
	ctx = context.WithValue(ctx, crudKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	if len(groupInt) > 0 {
		sql := basedboperat.ReflectModelPKJsonKey(model) + " in ("
		var args []interface{}
		for n, v := range groupInt {
			if n == 0 {
				sql = sql + "?"
			} else {
				sql = sql + ",?"
			}
			args = append(args, v)
		}
		sql = sql + ")"
		count, err := session.Delete(model, nil, sql, args...)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
		rpcConn.WriteResult(count)
		return
	}

	if len(groupString) > 0 {
		sql := basedboperat.ReflectModelPKJsonKey(model) + " in ("
		var args []interface{}
		for n, v := range groupString {
			if n == 0 {
				sql = sql + "?"
			} else {
				sql = sql + ",?"
			}
			args = append(args, v)
		}
		sql = sql + ")"
		count, err := session.Delete(model, nil, sql, args...)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
		rpcConn.WriteResult(count)
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

func (c *Crud) detail(rpcConn JsonRpcConnection) {
	modelType := c.modelObject
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	rpcConn.ReadParams(&model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey ContextKey = JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	var crudKey ContextKey = IS_CRUD
	ctx = context.WithValue(ctx, crudKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	var conditionStr string
	var conditionArgs []interface{}
	if len(c.DetailConditionFixed) > 0 {
		for _, f := range c.DetailConditionFixed {
			result, err := f(rpcConn)
			if err != nil {
				rpcConn.WriteError(500, err.Error())
				return
			}
			for n, condition := range result {
				if n == 0 {
					conditionStr = condition.Field + " " + condition.Operator + " ?"
				} else {
					conditionStr = conditionStr + " and " + condition.Field + " " + condition.Operator + " ?"
				}

				conditionArgs = append(conditionArgs, condition.Value)
			}
		}
	}

	err = session.Get(model, nil, conditionStr, conditionArgs...)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func (c *Crud) list(rpcConn JsonRpcConnection) {
	modelType := c.modelObject
	modelTypes := c.modelObjects
	model := reflect.New(reflect.TypeOf(modelType)).Interface()
	models := reflect.New(reflect.TypeOf(modelTypes)).Interface()

	var list basedboperat.List
	rpcConn.ReadParams(&list, &model)

	jsonrpcCtx := rpcConn.GetRpcRequest().Ctx
	ctx := context.Background()
	for k, v := range jsonrpcCtx {
		ctx = context.WithValue(ctx, k, v)
	}

	var contextKey ContextKey = JSON_RPC_CONNECTION_CONTEXT_KEY
	ctx = context.WithValue(ctx, contextKey, rpcConn)

	var crudKey ContextKey = IS_CRUD
	ctx = context.WithValue(ctx, crudKey, rpcConn)

	session, err := basedboperat.WithContext(ctx)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	if len(c.ListConditionFixed) > 0 {
		for _, f := range c.ListConditionFixed {
			conditions, err := f(rpcConn)
			if err != nil {
				rpcConn.WriteError(500, err.Error())
				return
			}
			list.Where = append(list.Where, conditions...)
		}
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

/*
reg jsonrpc method
if CRUD_FLAG is null ,Will Reg All Method
*/
func (c *Crud) AddInterceptor(f func(conn JsonRpcConnection) error, CRUD_FLAG ...string) {
	if CRUD_FLAG == nil {
		c.addInterceptor(f, CRUD_FLAG_CREATE, CRUD_FLAG_DELETE, CRUD_FLAG_DETAIL, CRUD_FLAG_LIST, CRUD_FLAG_UPDATE)
		return
	}
	c.addInterceptor(f, CRUD_FLAG...)
}

func (c *Crud) addInterceptor(f func(conn JsonRpcConnection) error, CRUD_FLAG ...string) {
	for _, v := range CRUD_FLAG {
		switch v {
		case CRUD_FLAG_CREATE:
			if c.CreateApi != nil {
				c.CreateApi.AddInterceptor(f)
			}

		case CRUD_FLAG_DELETE:
			if c.DeleteApi != nil {
				c.DeleteApi.AddInterceptor(f)
			}

		case CRUD_FLAG_DETAIL:
			if c.DetailApi != nil {
				c.DetailApi.AddInterceptor(f)
			}

		case CRUD_FLAG_LIST:
			if c.ListApi != nil {
				c.ListApi.AddInterceptor(f)
			}

		case CRUD_FLAG_UPDATE:
			if c.UpdateApi != nil {
				c.UpdateApi.AddInterceptor(f)
			}
		}
	}
}

func (c *Crud) regAPI(CRUD_FLAG ...string) {
	for _, v := range CRUD_FLAG {
		switch v {
		case CRUD_FLAG_CREATE:
			c.CreateApi = SetFunc(c.baseMethod+"/create", c.create)
		case CRUD_FLAG_DELETE:
			c.DeleteApi = SetFunc(c.baseMethod+"/delete", c.delete)
		case CRUD_FLAG_DETAIL:
			c.DetailApi = SetFunc(c.baseMethod+"/detail", c.detail)
		case CRUD_FLAG_LIST:
			c.ListApi = SetFunc(c.baseMethod+"/list", c.list)
		case CRUD_FLAG_UPDATE:
			c.UpdateApi = SetFunc(c.baseMethod+"/update", c.update)
		}
	}
}
