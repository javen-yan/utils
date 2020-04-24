package server

import (
	"errors"
	"fmt"
	"github.com/yanzongzhen/Logger/logger"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Service struct {
	Method      string
	Path        string
	Handler     interface{}
	ParamKey    string
	Interceptor []Interceptor
}

type JsonResult interface {
	ToJson() []byte
}

var serviceMap map[string]*Service

func init() {
	serviceMap = make(map[string]*Service)
}

func RegisterService(path string, service *Service) {
	rv := reflect.ValueOf(service.Handler)
	if rv.Kind() != reflect.Ptr {
		panic(errors.New("service handler type must be ptr"))
	}
	path = strings.ToLower(service.Method) + path
	serviceMap[path] = service
}

func handle(ctx *Context) {
	requestPath := ctx.Request.URL.Path
	res := strings.Split(requestPath, "/")
	p := strings.Join(res[:len(res)-1], "/")
	p = strings.ToLower(ctx.Request.Method) + p
	logger.Debug(serviceMap[p])
	s, ok := serviceMap[p]
	if !ok {
		ctx.Writer.WriteHeader(404)
		_, _ = ctx.Writer.Write([]byte("Not Found"))
		return
	}
	method := ctx.Params.ByName(s.ParamKey)
	rv := reflect.ValueOf(s.Handler)

	logger.Debug(capitalize(method))
	m := rv.MethodByName(capitalize(method))
	if !m.IsValid() || m.IsNil() || m.IsZero() {
		ctx.Writer.WriteHeader(404)
		_, _ = ctx.Writer.Write([]byte("Not Found " + method))
		return
	}

	if s.Interceptor != nil && len(s.Interceptor) > 0 {
		for _, i := range s.Interceptor {
			i(ctx)
		}
	}
	in := make([]reflect.Value, 0, 1)
	//in = append(in, reflect.ValueOf(request))
	//if !utils.IsEmpty(params.ByName("account")) {
	//	in = append(in, reflect.ValueOf(params))
	//}
	in = append(in, reflect.ValueOf(ctx))
	m.Call(in)
}

func startHttpServer(port int) {
	httpServer := &http.Server{
		Addr: ":" + strconv.Itoa(port),
	}
	r := httprouter.New()
	httpServer.Handler = r
	logger.Debug("START", len(serviceMap))
	logger.Debug(serviceMap)

	for _, h := range serviceMap {
		logger.Debug(h.Path)
		r.Handle(h.Method, h.Path, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
			defer func() {
				if request.Body != nil {
					_ = request.Body.Close()
				}
			}()
			defer func() {
				if r := recover(); r != nil {
					if res, ok := r.(JsonResult); ok {
						logger.Debug(string(res.ToJson()))
						writer.Header().Add("Content-Type", "application/json")
						_, _ = writer.Write(res.ToJson())
					} else if err, ok := r.(error); ok {
						writer.Header().Add("Content-Type", "application/json")
						logger.Debug(err)
						res := HttpRes{Code: ErrServer, Msg: err.Error()}
						_, _ = writer.Write(res.ToJson())
					} else {
						writer.Header().Add("Content-Type", "application/json")
						logger.Debug(r)
						logger.Debug(reflect.TypeOf(r).String())
						_, _ = writer.Write([]byte(fmt.Sprintf("{\"code\":%s,\"message\": \"%s\"}", "0001", "未知的错误")))
					}
				}
			}()
			ctx := NewContext(request, writer, params)

			//Auth(ctx)
			Time(handle)(ctx)
		})
	}
	go func() {
		err := httpServer.ListenAndServe()
		logger.Fatal(err)
	}()
}

//首字母大写
func capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
