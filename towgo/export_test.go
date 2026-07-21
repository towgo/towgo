package towgo

func ResetForTest() {
	middlewares = nil
	recoverHandler = DefaultRecoverHandler
	methods = map[string]*HandlerInfo{}
	funcs = map[string]*Api{}
	crudMap = map[string]*Crud{}
}
