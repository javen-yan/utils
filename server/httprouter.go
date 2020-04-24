package server

import (
	"github.com/yanzongzhen/Logger/logger"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

type Params = httprouter.Params

type IcityHttpHandler func(request *http.Request, params Params) ([]byte, error)

func StartHttpServer(port int, path string, handler IcityHttpHandler) {
	router := httprouter.New()

	router.POST(path, func(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {
		defer func() {
			if request.Body != nil {
				_ = request.Body.Close()
			}
		}()

		startTime := time.Now().UnixNano()
		logger.Infof("请求开始 ip:%s --请求地址:%s", request.RemoteAddr, request.URL.String())
		res, err := handler(request, ps)

		if err != nil {
			logger.Error(err)
		}
		//if err != nil {
		//	writeResponse(res, writer)
		//	closeFunc()
		//}
		logger.Debug(string(res))
		logger.Debug(len(res))
		writeResponse(res, writer)
		//writer.Write(res)
		logger.Infof("请求结束 --耗时:%d ms", (time.Now().UnixNano()-startTime)/int64(time.Millisecond))
	})
	logger.Debugf("http服务启动============port %d", port)
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + strconv.Itoa(port),
		WriteTimeout: time.Second * 15,
	}
	err := srv.ListenAndServe()
	if err != nil {
		logger.Error(err)
	}
}

func writeResponse(response []byte, writer http.ResponseWriter) {
	//var responseData []byte
	//var err error
	//if response != nil {
	//	responseData, err = json.Marshal(response)
	//	if err != nil {
	//		res := HttpRes{ErrServer, "未知异常", nil}
	//		responseData = res.ToJson()
	//	}
	//} else {
	//	res := HttpRes{ErrServer, "未知异常", nil}
	//	responseData = res.ToJson()
	//}
	writer.Header().Add("Content-Type", "application/json")
	_, _ = writer.Write(response)
	//logger.Debug(n)
	//logger.Error(err)
}
