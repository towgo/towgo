package terror_test

import (
	"errors"
	"fmt"
	"github.com/towgo/towgo/errors/tcode"
	"github.com/towgo/towgo/errors/terror"
	"log"
)

func ExampleNewCode() {

	defer func() {
		if exception := recover(); exception != nil {
			if v, ok := exception.(error); ok && terror.HasStack(v) {
				log.Printf("err = %+v\n", v)
			} else {
				log.Printf("err = %+v\n", terror.NewCodef(tcode.CodeInternalPanic, "exception recovered: %+v", exception))
			}
		}
	}()
	err := terror.NewCode(tcode.New(10000, "", nil), "My Error")
	//panic(err)
	fmt.Println(err.Error())
	fmt.Println(terror.Code(err))
	log.Println(terror.Stack(err))
	log.Printf("err = %+v\n", err)
	// Output:
	// My Error
	// 10000
}

func ExampleNewCodef() {
	err := terror.NewCodef(tcode.New(10000, "", nil), "It's %s", "My Error")
	fmt.Println(err.Error())
	fmt.Println(terror.Code(err).Code())

	// Output:
	// It's My Error
	// 10000
}

func ExampleWrapCode() {
	err1 := errors.New("permission denied")
	err2 := terror.WrapCode(tcode.New(10000, "", nil), err1, "Custom Error")
	fmt.Println(err2.Error())
	fmt.Println(terror.Code(err2).Code())

	// Output:
	// Custom Error: permission denied
	// 10000
}

func ExampleWrapCodef() {
	err1 := errors.New("permission denied")
	err2 := terror.WrapCodef(tcode.New(10000, "", nil), err1, "It's %s", "Custom Error")
	fmt.Println(err2.Error())
	fmt.Println(terror.Code(err2).Code())

	// Output:
	// It's Custom Error: permission denied
	// 10000
}

func ExampleEqual() {
	err1 := errors.New("permission denied")
	err2 := terror.New("permission denied")
	err3 := terror.NewCode(tcode.CodeNotAuthorized, "permission denied")
	fmt.Println(terror.Equal(err1, err2))
	fmt.Println(terror.Equal(err2, err3))

	// Output:
	// true
	// false
}

func ExampleIs() {
	err1 := errors.New("permission denied")
	err2 := terror.Wrap(err1, "operation failed")
	fmt.Println(terror.Is(err1, err1))
	fmt.Println(terror.Is(err2, err2))
	fmt.Println(terror.Is(err2, err1))
	fmt.Println(terror.Is(err1, err2))

	// Output:
	// true
	// true
	// true
	// false
}

func ExampleCode() {
	err1 := terror.NewCode(tcode.CodeInternalError, "permission denied")
	err2 := terror.Wrap(err1, "operation failed")
	fmt.Println(terror.Code(err1))
	fmt.Println(terror.Code(err2))

	// Output:
	// 50:Internal Error
	// 50:Internal Error
}

func ExampleHasCode() {
	err1 := terror.NewCode(tcode.CodeInternalError, "permission denied")
	err2 := terror.Wrap(err1, "operation failed")
	fmt.Println(terror.HasCode(err1, tcode.CodeOK))
	fmt.Println(terror.HasCode(err2, tcode.CodeInternalError))

	// Output:
	// false
	// true
}
