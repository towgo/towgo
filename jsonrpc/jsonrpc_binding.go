package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gstructs"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gmeta"
	"github.com/gogf/gf/v2/util/gtag"
	"github.com/gogf/gf/v2/util/gvalid"
)

var (
	// requestPtrType identifies handlers that accept *Request directly.
	requestPtrType = reflect.TypeOf((*Request)(nil))
	// connectionType identifies handlers using the JsonRpcConnection API.
	connectionType = reflect.TypeOf((*JsonRpcConnection)(nil)).Elem()
	// contextType identifies GoFrame-style handlers with context.Context.
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	// errorType identifies handler return values implementing error.
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

// bind registers one function, struct value, or controller object on the group.
func (g *RouterGroup) bind(item any) {
	reflectValue := reflect.ValueOf(item)
	if !reflectValue.IsValid() {
		return
	}
	reflectType := reflectValue.Type()

	if reflectType.Kind() == reflect.Func {
		handler, path, err := createHandler(reflectValue)
		if err != nil {
			panic(err)
		}
		g.Handle(path, handler)
		return
	}

	if reflectType.Kind() == reflect.Struct {
		newValue := reflect.New(reflectType)
		newValue.Elem().Set(reflectValue)
		reflectValue = newValue
		reflectType = reflectValue.Type()
	}
	if reflectType.Kind() != reflect.Ptr || reflectType.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("invalid bind object: %s", reflectType.String()))
	}

	for i := 0; i < reflectValue.NumMethod(); i++ {
		method := reflectType.Method(i)
		handler, path, err := createHandler(reflectValue.Method(i))
		if err != nil {
			continue
		}
		if path == "" || path == "/" {
			path = nameToMethod(method.Name)
		}
		g.Handle(path, handler)
	}
}

// createHandler converts supported function signatures to HandlerFunc.
func createHandler(value reflect.Value) (HandlerFunc, string, error) {
	if !value.IsValid() || value.Kind() != reflect.Func {
		return nil, "", gerror.NewCode(gcode.CodeInvalidParameter, "handler should be func")
	}

	if handler, ok := value.Interface().(HandlerFunc); ok {
		return handler, "/", nil
	}

	t := value.Type()
	if t.NumIn() == 1 && t.In(0) == requestPtrType && t.NumOut() == 0 {
		return func(r *Request) {
			value.Call([]reflect.Value{reflect.ValueOf(r)})
		}, "/", nil
	}
	if t.NumIn() == 1 && t.In(0).Implements(connectionType) && t.NumOut() == 0 {
		return func(r *Request) {
			value.Call([]reflect.Value{reflect.ValueOf(r.AsConnection())})
		}, "/", nil
	}

	if t.NumIn() != 2 || t.NumOut() != 2 {
		return nil, "", gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid handler "%s", should be func(*towgo.Request) or func(context.Context, *BizReq)(*BizRes, error)`,
			t.String(),
		)
	}
	if !t.In(0).Implements(contextType) {
		return nil, "", gerror.NewCodef(gcode.CodeInvalidParameter, `invalid handler "%s", first input should be context.Context`, t.String())
	}
	if t.In(1).Kind() != reflect.Ptr || t.In(1).Elem().Kind() != reflect.Struct {
		return nil, "", gerror.NewCodef(gcode.CodeInvalidParameter, `invalid handler "%s", second input should be pointer to struct`, t.String())
	}
	if !t.Out(1).Implements(errorType) {
		return nil, "", gerror.NewCodef(gcode.CodeInvalidParameter, `invalid handler "%s", second output should be error`, t.String())
	}
	if t.Out(0).Kind() != reflect.Ptr || t.Out(0).Elem().Kind() != reflect.Struct {
		return nil, "", gerror.NewCodef(gcode.CodeInvalidParameter, `invalid handler "%s", first output should be pointer to struct`, t.String())
	}

	reqPtr := reflect.New(t.In(1).Elem()).Interface()
	path := gmeta.Get(reqPtr, gtag.Path).String()
	fields, err := gstructs.Fields(gstructs.FieldsInput{
		Pointer:         reqPtr,
		RecursiveOption: gstructs.RecursiveOptionEmbedded,
	})
	if err != nil {
		return nil, "", err
	}

	return func(r *Request) {
		input := reflect.New(t.In(1).Elem())
		if err := parseParams(r, fields, input.Interface()); err != nil {
			r.WithError(err)
			return
		}

		results := value.Call([]reflect.Value{
			reflect.ValueOf(r.Context()),
			input,
		})
		if !results[1].IsNil() {
			if err, ok := results[1].Interface().(error); ok {
				r.WithError(err)
			}
			return
		}
		if !results[0].IsNil() {
			r.SetResult(results[0].Interface())
		}
	}, path, nil
}

// parseParams converts JSON-RPC params into a GoFrame-style request object.
func parseParams(r *Request, fields []gstructs.Field, pointer any) error {
	var data map[string]any
	paramsBytes, err := json.Marshal(r.Params())
	if err != nil {
		return gerror.Wrap(err, "marshal params failed")
	}
	if len(paramsBytes) == 0 || string(paramsBytes) == "null" {
		paramsBytes = []byte("{}")
	}
	if err := json.Unmarshal(paramsBytes, &data); err != nil {
		return gerror.Wrap(err, "unmarshal params failed")
	}
	mergeDefaultStructValue(fields, data)
	if err := gconv.Structs(data, pointer); err != nil {
		return gerror.Wrap(err, "invalid parameter")
	}
	if err := gvalid.New().Bail().Data(pointer).Assoc(data).Run(r.Context()); err != nil {
		return gerror.New(err.Error())
	}
	return nil
}

// mergeDefaultStructValue applies d/default tags before validation.
func mergeDefaultStructValue(fields []gstructs.Field, data map[string]any) {
	for _, field := range fields {
		if tagValue := field.TagDefault(); tagValue != "" {
			if foundKey, foundValue := findPossibleItemByKey(data, field.Name()); foundKey == "" || foundValue == nil {
				data[field.Name()] = tagValue
			}
		}
	}
}

// nameToMethod converts an exported Go method name to a JSON-RPC method path.
func nameToMethod(name string) string {
	if name == "" {
		return "/"
	}
	part := strings.Builder{}
	if gstr.IsLetterUpper(name[0]) {
		part.WriteByte(name[0] + 32)
	} else {
		part.WriteByte(name[0])
	}
	part.WriteString(name[1:])
	return "/" + part.String()
}

// findPossibleItemByKey finds a map item with case-insensitive key matching.
func findPossibleItemByKey(data map[string]any, key string) (string, any) {
	for k, v := range data {
		if strings.EqualFold(k, key) {
			return k, v
		}
	}
	return "", nil
}
