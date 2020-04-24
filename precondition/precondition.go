package precondition

import (
	"errors"
	"github.com/yanzongzhen/utils"
	"reflect"
)

//检查参数是否合法，只能检查 指针类型和字符串类型，如果是
func CheckArgValid(error string, arg interface{}) {
	rvArg := reflect.ValueOf(arg)

	switch rvArg.Kind() {
	case reflect.Ptr:
		if rvArg.IsNil() {
			panic(errors.New(error))
		}
		break
	case reflect.String:
		if utils.IsEmpty(arg.(string)) {
			panic(errors.New(error))
		}
		break
	default:
		panic(errors.New("不支持的类型"))
	}
}

func CheckArgsValid(error string, args ...interface{}) {
	for _, a := range args {
		CheckArgValid(error, a)
	}
}
