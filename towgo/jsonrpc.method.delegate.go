/*
json rpc 2.0 方法委托
by:liangliangit
version 2.2
*/
package towgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gstructs"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gmeta"
	"github.com/gogf/gf/v2/util/gtag"
	"github.com/gogf/gf/v2/util/gutil"
	"github.com/gogf/gf/v2/util/gvalid"
	"github.com/towgo/towgo/errors/tcode"
	"github.com/towgo/towgo/errors/terror"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/towgo/towgo/lib/system"
)

// 委托任务列表
// var funcs map[string]func(JsonRpcConnection) = map[string]func(JsonRpcConnection){}
var lock sync.Mutex
var funcs map[string]*Api = map[string]*Api{}
var lockedMethods sync.Map

const (
	parseTypeRequest = iota
	parseTypeQuery
	parseTypeForm
)

var BeforExec func(rpcConn JsonRpcConnection)
var AfterExec func(rpcConn JsonRpcConnection)
var DefaultExec func(rpcConn JsonRpcConnection) = func(rpcConn JsonRpcConnection) {
	if exception := recover(); exception != nil {
		if v, ok := exception.(error); ok && terror.HasStack(v) {
			log.Printf("err %+v \n", v)
		} else {
			log.Printf("recover exception %+v\n", terror.NewCodef(tcode.CodeInternalPanic, "%+v", exception))
		}
		rpcConn.WriteError(500, DEFAULT_ERROR_MSG)
	}
}
var OnMethodNotFound func(rpcConn JsonRpcConnection)

var Execmap map[string]int

// 接口对象
type Api struct {
	method              string
	f                   func(JsonRpcConnection)
	interceptorHandller []func(conn JsonRpcConnection) error
}

func (a *Api) Method() string {
	return a.method
}

func (a *Api) Exec(rpcConn JsonRpcConnection) {

	//运行拦截器
	err := a.interceptor(rpcConn)
	if err != nil {
		rpcConn.WriteError(500, err.Error())
		return
	}

	//运行方法
	a.f(rpcConn)
}

func (a *Api) AddInterceptor(args ...func(conn JsonRpcConnection) error) {
	a.interceptorHandller = append(a.interceptorHandller, args...)
}

func (a *Api) interceptor(rpcConn JsonRpcConnection) error {
	for _, v := range a.interceptorHandller {
		err := v(rpcConn)
		if err != nil {
			return err
		}
	}
	return nil
}

// 查询method是否存在
func HasMethod(method string) bool {
	_, ok := funcs[method]
	return ok
}

// 为所有method增加头
func AddMethodHead(methodHead string) {
	lock.Lock()
	defer lock.Unlock()
	var newmap map[string]*Api = map[string]*Api{}
	for k, v := range funcs {
		newmap[methodHead+k] = v
	}
	funcs = newmap
}

// 获取method列表
func GetMethods() (method []string) {
	for k := range funcs {
		method = append(method, k)
	}
	return
}

func http_jsonrpc_wrapper(w http.ResponseWriter, r *http.Request) {

	urlPath := r.URL.Path
	rpcRequest := NewJsonrpcrequest()
	rpcRequest.Method = urlPath
	rpcRequest.Session = r.Header.Get("session")

	conn := &HttpRpcConnection{
		guid:        "HTTP:" + system.GetGUID().Hex(),
		response:    w,
		request:     r,
		rpcRequest:  rpcRequest,
		rpcResponse: NewJsonrpcresponse(),
		httpwrapper: true,
	}

	err := defaultJsonRpcInterceptor(conn)
	if err != nil {
		conn.isConnected = false
		log.Print(err.Error())
		return //拦截后 rpc响应由拦截器处理，  不需要再次响应
	}
	Exec(conn)
}

// 将jsonrpc method路由接口兼容为HTTP路由接口 兼容restful风格
func MethodToHttpPathInterface(serveMux *http.ServeMux) {
	for k := range funcs {
		method := "/" + strings.TrimLeft(k, "/")
		serveMux.HandleFunc(method, http_jsonrpc_wrapper)
	}
}

// 锁定指定method （可用于许可证到期锁定相关服务）
func MethodLock(method string) {
	lockedMethods.Store(method, "")
}

// 解锁method
func MethodUnlock(method string) {
	lockedMethods.Delete(method)
}

// 锁定所有method （可用于许可证到期锁定相关服务,排除关键性业务不锁定）
func MethodLockAll(excludeMethods ...string) {
	for k := range funcs {
		lockedMethods.Store(k, "")
	}
	for _, v := range excludeMethods {
		lockedMethods.Delete(v)
	}
}

// 解锁所有method
func MethodUnlockAll(excludeMethods ...string) {
	lockedMethods.Range(func(key, _ any) bool {
		lockedMethods.Delete(key)
		return true
	})
	for _, v := range excludeMethods {
		lockedMethods.Store(v, "")
	}
}

// 设定委托任务
func SetFunc(method string, f func(JsonRpcConnection)) *Api {
	lock.Lock()
	defer lock.Unlock()
	api := &Api{
		method: method,
		f:      f,
	}
	funcs[method] = api
	return api
}
func RemoveFunc(method string) {
	lock.Lock()
	defer lock.Unlock()
	delete(funcs, method)
}

// 委托执行任务
func Exec(rpcConn JsonRpcConnection) {
	if BeforExec != nil {
		BeforExec(rpcConn)
	}
	rpcResponse := rpcConn.GetRpcResponse()
	rpcRequest := rpcConn.GetRpcRequest()
	if rpcRequest == nil {
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	if rpcRequest.Method == "" {
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	//判断是否锁定
	_, ok := lockedMethods.Load(rpcRequest.Method)
	if ok {
		rpcResponse.Error.Set(500, "Method has been locked!")
		rpcConn.Write()
		return
	}

	api, exists := funcs[rpcRequest.Method]
	// 判断委托是否存在
	if !exists {
		//如果注册了Method not found处理函数  不进行默认响应
		if OnMethodNotFound != nil {
			OnMethodNotFound(rpcConn)
			return
		}
		//默认响应 Method not found找不到方法
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	// 执行委托的程序
	api.Exec(rpcConn)
	if AfterExec != nil {
		AfterExec(rpcConn)
	}
}

func BindFunc(method string, f func(JsonRpcConnection)) {
	SetFunc(method, f)
}
func BindObject(method string, object interface{}) {
	var (
		reflectValue = reflect.ValueOf(object)
		reflectType  = reflectValue.Type()
	)
	// If given `object` is not pointer, it then creates a temporary one,
	// of which the value is `reflectValue`.
	// It then can retrieve all the methods both of struct/*struct.
	if reflectValue.Kind() == reflect.Struct {
		newValue := reflect.New(reflectType)
		newValue.Elem().Set(reflectValue)
		reflectValue = newValue
		reflectType = reflectValue.Type()
	}
	structName := reflectType.Elem().Name()
	pkgPath := reflectType.Elem().PkgPath()
	pkgName := gfile.Basename(pkgPath)
	var methodMap map[string]bool
	for i := 0; i < reflectValue.NumMethod(); i++ {
		methodName := reflectType.Method(i).Name
		if methodMap != nil && !methodMap[methodName] {
			continue
		}

		objName := gstr.Replace(reflectType.String(), fmt.Sprintf(`%s.`, pkgName), "")
		if objName[0] == '*' {
			objName = fmt.Sprintf(`(%s)`, objName)
		}

		funcInfo, err := checkAndCreateFuncInfo(reflectValue.Method(i).Interface(), pkgPath, objName, methodName)
		if err != nil {
			panic(err)
		}
		uri := mergeBuildInNameToPattern(method, structName, methodName, true)
		if funcInfo.Path != "" {
			split := strings.Split(funcInfo.Path, "/")
			replace := strings.Replace(method, "/", "", 1)
			if replace != split[1] {
				uri = method + funcInfo.Path
			} else {
				uri = funcInfo.Path
			}
		}
		SetFunc(uri, funcInfo.Func)
	}

}
func nameToUri(name string) string {
	part := bytes.NewBuffer(nil)
	if gstr.IsLetterUpper(name[0]) {
		part.WriteByte(name[0] + 32)
	} else {
		part.WriteByte(name[0])
	}
	part.WriteString(name[1:])
	return part.String()
}
func mergeBuildInNameToPattern(pattern string, structName, methodName string, allowAppend bool) string {
	structName = nameToUri(structName)
	methodName = nameToUri(methodName)
	pattern = strings.ReplaceAll(pattern, "{.struct}", structName)
	if strings.Contains(pattern, "{.method}") {
		return strings.ReplaceAll(pattern, "{.method}", methodName)
	}
	if !allowAppend {
		return pattern
	}
	// Check domain parameter.
	var (
		array = strings.Split(pattern, "@")
		uri   = strings.TrimRight(array[0], "/") + "/" + methodName
	)
	// Append the domain parameter to URI.
	if len(array) > 1 {
		return uri + "@" + array[1]
	}
	return uri
}

// HandlerFunc is request handler function.
type HandlerFunc = func(JsonRpcConnection)

// handlerFuncInfo contains the HandlerFunc address and its reflection type.
type handlerFuncInfo struct {
	Func            HandlerFunc      // Handler function address.
	Type            reflect.Type     // Reflect type information for current handler, which is used for extensions of the handler feature.
	Value           reflect.Value    // Reflect value information for current handler, which is used for extensions of the handler feature.
	IsStrictRoute   bool             // Whether strict route matching is enabled.
	ReqStructFields []gstructs.Field // Request struct fields.
	Path            string
}

func checkAndCreateFuncInfo(
	f interface{}, pkgPath, structName, methodName string,
) (funcInfo handlerFuncInfo, err error) {
	funcInfo = handlerFuncInfo{
		Type:  reflect.TypeOf(f),
		Value: reflect.ValueOf(f),
	}
	if handlerFunc, ok := f.(HandlerFunc); ok {
		funcInfo.Func = handlerFunc

		return
	}

	var (
		reflectType    = funcInfo.Type
		inputObject    reflect.Value
		inputObjectPtr interface{}
	)
	if reflectType.NumIn() != 2 || reflectType.NumOut() != 2 {
		if pkgPath != "" {
			err = gerror.NewCodef(
				gcode.CodeInvalidParameter,
				`invalid handler: %s.%s.%s defined as "%s", but "func(JsonRpcConnection)" or "func(context.Context, *BizReq)(*BizRes, error)" is required`,
				pkgPath, structName, methodName, reflectType.String(),
			)
		} else {
			err = gerror.NewCodef(
				gcode.CodeInvalidParameter,
				`invalid handler: defined as "%s", but "func(JsonRpcConnection)" or "func(context.Context, *BizReq)(*BizRes, error)" is required`,
				reflectType.String(),
			)
		}
		return
	}

	if !reflectType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		err = gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid handler: defined as "%s", but the first input parameter should be type of "context.Context"`,
			reflectType.String(),
		)
		return
	}

	if !reflectType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		err = gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid handler: defined as "%s", but the last output parameter should be type of "error"`,
			reflectType.String(),
		)
		return
	}

	if reflectType.In(1).Kind() != reflect.Ptr ||
		(reflectType.In(1).Kind() == reflect.Ptr && reflectType.In(1).Elem().Kind() != reflect.Struct) {
		err = gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid handler: defined as "%s", but the second input parameter should be type of pointer to struct like "*BizReq"`,
			reflectType.String(),
		)
		return
	}

	// Do not enable this logic, as many users are already using none struct pointer type
	// as the first output parameter.
	/*
		if reflectType.Out(0).Kind() != reflect.Ptr ||
			(reflectType.Out(0).Kind() == reflect.Ptr && reflectType.Out(0).Elem().Kind() != reflect.Struct) {
			err = gerror.NewCodef(
				gcode.CodeInvalidParameter,
				`invalid handler: defined as "%s", but the first output parameter should be type of pointer to struct like "*BizRes"`,
				reflectType.String(),
			)
			return
		}
	*/

	funcInfo.IsStrictRoute = true

	inputObject = reflect.New(funcInfo.Type.In(1).Elem())
	inputObjectPtr = inputObject.Interface()
	gmetaPath := gmeta.Get(inputObjectPtr, gtag.Path).String()

	// It retrieves and returns the request struct fields.
	fields, err := gstructs.Fields(gstructs.FieldsInput{
		Pointer:         inputObjectPtr,
		RecursiveOption: gstructs.RecursiveOptionEmbedded,
	})
	if err != nil {
		return funcInfo, err
	}
	funcInfo.Path = gmetaPath
	funcInfo.ReqStructFields = fields
	funcInfo.Func = createRouterFunc(funcInfo, funcInfo.ReqStructFields)
	return
}
func createRouterFunc(funcInfo handlerFuncInfo, fields []gstructs.Field) func(conn JsonRpcConnection) {
	return func(conn JsonRpcConnection) {
		var (
			ok          bool
			err         error
			inputValues = []reflect.Value{
				reflect.ValueOf(conn.GetRpcRequest().ctx),
			}
		)
		if funcInfo.Type.NumIn() == 2 {
			var inputObject reflect.Value
			if funcInfo.Type.In(1).Kind() == reflect.Ptr {
				inputObject = reflect.New(funcInfo.Type.In(1).Elem())
				err = doParse(conn, fields, inputObject.Interface(), parseTypeRequest)

			} else {
				inputObject = reflect.New(funcInfo.Type.In(1).Elem()).Elem()
				err = doParse(conn, fields, inputObject.Addr().Interface(), parseTypeRequest)
			}
			if err != nil {
				conn.WriteError(500, err.Error())
				return
			}
			inputValues = append(inputValues, inputObject)
		}
		// Call handler with dynamic created parameter values.
		results := funcInfo.Value.Call(inputValues)
		switch len(results) {
		case 1:
			if !results[0].IsNil() {
				if err, ok = results[0].Interface().(error); ok {
					conn.WriteError(500, err.Error())
					return
				}
			}

		case 2:
			if !results[1].IsNil() {
				if err, ok = results[1].Interface().(error); ok {
					conn.WriteError(500, err.Error())
					return
				}
			}
			result := results[0].Interface()
			conn.WriteResult(result)
			return
		}
	}
}

// doParse parses the request data to struct/structs according to request type.
func doParse(conn JsonRpcConnection, fields []gstructs.Field, pointer interface{}, requestType int) error {
	var (
		reflectVal1  = reflect.ValueOf(pointer)
		reflectKind1 = reflectVal1.Kind()
	)
	if reflectKind1 != reflect.Ptr {
		return gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid parameter type "%v", of which kind should be of *struct/**struct/*[]struct/*[]*struct, but got: "%v"`,
			reflectVal1.Type(),
			reflectKind1,
		)
	}
	var (
		reflectVal2  = reflectVal1.Elem()
		reflectKind2 = reflectVal2.Kind()
	)
	switch reflectKind2 {
	// Single struct, post content like:
	// 1. {"id":1, "name":"john"}
	// 2. ?id=1&name=john
	case reflect.Ptr, reflect.Struct:
		var (
			err  error
			data map[string]interface{}
		)
		// Converting.
		switch requestType {

		default:
			err = conn.ReadParams(&data)

			if err != nil {
				return err
			}
			err = mergeDefaultStructValue(fields, data, requestType)
			if err != nil {
				return err
			}
			err = gconv.Structs(data, pointer)
			if err != nil {
				return err
			}

		}
		// Validation.
		if err = gvalid.New().
			Bail().
			Data(pointer).
			Assoc(data).
			Run(conn.GetRpcRequest().ctx); err != nil {
			return err
		}

	// Multiple struct, it only supports JSON type post content like:
	// [{"id":1, "name":"john"}, {"id":, "name":"smith"}]
	case reflect.Array, reflect.Slice:
		// If struct slice conversion, it might post JSON/XML/... content,
		// so it uses `gjson` for the conversion.
		marshal, err := json.Marshal(conn.GetRpcRequest().Params)
		if err != nil {
			return err
		}
		j, err := gjson.LoadContent(marshal)
		if err != nil {
			return err
		}
		if err = j.Var().Scan(pointer); err != nil {
			return err
		}
		for i := 0; i < reflectVal2.Len(); i++ {
			if err = gvalid.New().
				Bail().
				Data(reflectVal2.Index(i)).
				Assoc(j.Get(gconv.String(i)).Map()).
				Run(conn.GetRpcRequest().ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// mergeDefaultStructValue merges the request parameters with default values from struct tag definition.
func mergeDefaultStructValue(fields []gstructs.Field, data map[string]interface{}, pointer interface{}) error {

	if len(fields) > 0 {
		for _, field := range fields {
			if tagValue := field.TagDefault(); tagValue != "" {
				mergeTagValueWithFoundKey(data, false, field.Name(), field.Name(), tagValue)
			}
		}
		return nil
	}

	// provide non strict routing
	tagFields, err := gstructs.TagFields(pointer, []string{gtag.DefaultShort, gtag.Default})
	if err != nil {
		return err
	}
	if len(tagFields) > 0 {
		for _, field := range tagFields {
			mergeTagValueWithFoundKey(data, false, field.Name(), field.Name(), field.TagValue)
		}
	}

	return nil
}

// mergeTagValueWithFoundKey merges the request parameters when the key does not exist in the map or overwritten is true or the value is nil.
func mergeTagValueWithFoundKey(data map[string]interface{}, overwritten bool, findKey string, fieldName string, tagValue interface{}) {
	if foundKey, foundValue := gutil.MapPossibleItemByKey(data, findKey); foundKey == "" {
		data[fieldName] = tagValue
	} else {
		if overwritten || foundValue == nil {
			data[foundKey] = tagValue
		}
	}
}
