package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/yanzongzhen/Logger/logger"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var httpClient *http.Client

type HttpResponse struct {
	httpStatus     int
	responseHeader http.Header
	body           []byte
	err            error
}

func GetInitHttpResponse(status int, header http.Header, body []byte, err error) *HttpResponse {
	return &HttpResponse{status, header, body, err}
}

func (res *HttpResponse) SetStatus(status int) {
	res.httpStatus = status
}

func (res *HttpResponse) SetHeader(header http.Header) {
	res.responseHeader = header
}

func (res *HttpResponse) SetBody(body []byte) {
	res.body = body
}

func (res *HttpResponse) SetErr(err error) {
	res.err = err
}

//func InitHttpResponse() (*HttpResponse) {
//
//}

func (res *HttpResponse) GetStatus() int {
	return res.httpStatus
}

func (res *HttpResponse) GetHeader() http.Header {
	return res.responseHeader
}

func (res *HttpResponse) GetBody() []byte {
	return res.body
}

func (res *HttpResponse) GetErr() error {
	return res.err
}

func init() {
	InitHttpClientWithTimeOut(time.Second * 120)
}

func InitHttpClient(certPath ...string) {
	pool := x509.NewCertPool()
	for _, caCertPath := range certPath {
		caCrt, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			fmt.Println("ReadFile err:", err)
			return
		}
		pool.AppendCertsFromPEM(caCrt)
	}

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{RootCAs: pool},
		DisableCompression: true,
	}
	httpClient = &http.Client{
		Transport: tr,
	}
}

func InitHttpClientWithTimeOut(timeOutDuration time.Duration) {
	InitHttpClientWithTimeOutAndCert(timeOutDuration, nil)
}

func InitHttpClientWithTimeOutAndCert(timeOutDuration time.Duration, certs []tls.Certificate) {
	var tr *http.Transport
	if certs != nil {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, Certificates: certs},
		}
	} else {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	//cookieJar, _ := cookiejar.New(nil)
	httpClient = &http.Client{
		Transport: tr,
		Timeout:   timeOutDuration,
		//CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//	return nil
		//},
		//
	}
}

func SendHttpRequest(request *http.Request) *HttpResponse {
	return sendHttpRequest(request)
}

func PostNotKeepAliveWithTimeOut(url string, header map[string]string, body []byte, timeOut string) *HttpResponse {
	return post(url, header, body, timeOut, true)
}
func PostNotKeepAlive(url string, header map[string]string, body []byte) *HttpResponse {
	return post(url, header, body, "", true)
}
func PostWithTimeOut(url string, header map[string]string, body []byte, timeOut string) *HttpResponse {
	return post(url, header, body, timeOut, false)
}

func Post(url string, header map[string]string, body []byte) *HttpResponse {
	return post(url, header, body, "", false)
}

func Put(url string, header map[string]string, body []byte) *HttpResponse {
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	logger.Debugf("请求体为:%v", string(body))
	if err != nil {
		logger.Errorf("生成request失败:%v", err)
		return &HttpResponse{-1, nil, nil, err}
	}
	if header != nil {
		for k, v := range header {
			//request.Header.Add(k, v)
			request.Header[k] = []string{v}
		}
		logger.Debugf("请求头为:%v", request.Header)
	}
	return sendHttpRequest(request)
}
func GetNotKeepAliveWithTimeOut(requestUrl string, header map[string]string, params map[string]string, timeOut string) *HttpResponse {
	return get(requestUrl, header, params, timeOut, true) //11111
	//return platformGet(requestUrl, header, params, timeOut, true,convertMethod)//11111
}
func GetNotKeepAlive(requestUrl string, header map[string]string, params map[string]string) *HttpResponse {
	return get(requestUrl, header, params, "", true)
}

func GetWithTimeOut(requestUrl string, header map[string]string, params map[string]string, timeOut string) *HttpResponse {
	return get(requestUrl, header, params, timeOut, false) //111
	//return platformGet(requestUrl, header, params, timeOut, false)//111
}

func Get(requestUrl string, header map[string]string, params map[string]string) *HttpResponse {
	return get(requestUrl, header, params, "", false)
}

func get(requestUrl string, header map[string]string, params map[string]string, timeOut string, isNotKeepAlive bool) *HttpResponse {
	values := url.Values{}
	if params != nil {
		for k, v := range params {
			values.Add(k, v)
		}
	}
	logger.Debugf("requestUrl:%v", requestUrl)
	if len(values) > 0 {
		if strings.Contains(requestUrl, "?") {
			requestUrl = requestUrl + "&" + values.Encode()
		} else {
			requestUrl = requestUrl + "?" + values.Encode()
		}
	}
	logger.Debugf("拼装参数后的请求地址为:%v", requestUrl)
	logger.Debugf("请求体为:%v", values.Encode())
	request, err := http.NewRequest("GET", requestUrl, http.NoBody)
	if err != nil {
		logger.Errorf("生成request失败:%v", err)
		return &HttpResponse{-1, nil, nil, err}
	}
	if isNotKeepAlive {
		request.Close = true
	}
	if header != nil {

		for k, v := range header {
			//request.Header.Add(k, v)
			//request.Header[k] = []string{v}
			request.Header[k] = append(request.Header[k], v)
		}
		logger.Debugf("请求头为:%v", request.Header)
	}
	logger.Debugf("超时时间为:%v", timeOut)
	if !IsEmpty(timeOut) {
		ctx, cancel, err := getContextByTimeOut(timeOut)
		if err != nil {
			logger.Errorf("解析超时时间失败:%v", err)
			return &HttpResponse{-1, nil, nil, err}
		}
		defer cancel()
		request = request.WithContext(ctx)
	}
	return sendHttpRequest(request)
}

//func platformGet(requestUrl string, header map[string]string, params map[string]string, timeOut string, isNotKeepAlive bool,convertMethod interface{}) *HttpResponse {
//
//	Response := &HttpResponse{-1, nil, nil, errors.New("out of time")}
//	values := url.Values{}
//	resCh := make(chan *HttpResponse)
//	if params != nil {
//		for k, v := range params {
//			values.Add(k, v)
//		}
//	}
//	logger.Debugf("requestUrl:%v", requestUrl)
//	if len(values) > 0 {
//		if strings.Contains(requestUrl, "?") {
//			requestUrl = requestUrl + "&" + values.Encode()
//		} else {
//			requestUrl = requestUrl + "?" + values.Encode()
//		}
//	}
//	logger.Debugf("拼装参数后的请求地址为:%v", requestUrl)
//	logger.Debugf("请求体为:%v", values.Encode())
//	request, err := http.NewRequest("GET", requestUrl, http.NoBody)
//	if err != nil {
//		logger.Errorf("生成request失败:%v", err)
//		return &HttpResponse{-1, nil, nil, err}
//	}
//	if isNotKeepAlive {
//		request.Close = true
//	}
//	if header != nil {
//		for k, v := range header {
//			//request.Header.Add(k, v)
//			//request.Header[k] = []string{v}
//			request.Header[k] = append(request.Header[k], v)
//		}
//		logger.Debugf("请求头为:%v", request.Header)
//	}
//	logger.Debugf("超时时间为:%v", timeOut)
//	//if !IsEmpty(timeOut) {
//	//	ctx, cancel, err := getContextByTimeOut(timeOut)
//	//	if err != nil {
//	//		logger.Errorf("解析超时时间失败:%v", err)
//	//		return &HttpResponse{-1, nil, nil, err}
//	//	}
//	//	defer cancel()
//	//	request = request.WithContext(ctx)
//	//}
//
//	timeOutInt,_ := strconv.Atoi(timeOut)
//	timer := time.NewTimer(time.Duration(timeOutInt)*time.Second)
//
//	go platformSendHttpRequest(request,timer,resCh,convertMethod)
//
//
//	select {
//	case <-timer.C:
//		break
//	case Response = <-resCh:
//		break
//	}
//
//	return Response
//
//
//
//
//
//}
//
//func platformSendHttpRequest(request *http.Request,timer *time.Timer,resCh chan *HttpResponse,convertMethod interface{})  {
//	Response := sendHttpRequest(request)
//	isStop := timer.Stop()
//	if isStop {
//		resCh <- Response
//	}else {
//		switch convertMethod.(type) {
//		case :
//
//		}
//
//
//
//
//
//
//	}
//
//}

func post(url string, header map[string]string, body []byte, timeOut string, isNotKeepAlive bool) *HttpResponse {
	logger.Debugf("接口超时时间为:%v", timeOut)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	logger.Debugf("请求体为:%v", string(body))
	if err != nil {
		logger.Errorf("生成request失败:%v", err)
		return &HttpResponse{-1, nil, nil, err}
	}
	if isNotKeepAlive {
		request.Close = true
	}

	if header != nil {
		for k, v := range header {
			//request.Header[k] = []string{v}
			request.Header[k] = append(request.Header[k], v)
			//request.Header.Add(k,v)
		}
		logger.Debugf("请求头为:%v", request.Header)
	}
	if !IsEmpty(timeOut) {
		ctx, cancel, err := getContextByTimeOut(timeOut)
		if err != nil {
			logger.Errorf("解析超时时间失败:%v", err)
			return &HttpResponse{-1, nil, nil, err}
		}
		defer cancel()
		request = request.WithContext(ctx)
	}
	return sendHttpRequest(request)
}

func getContextByTimeOut(timeOut string) (context.Context, context.CancelFunc, error) {
	logger.Debugf("接口进入超时")
	t, err := DealTime(timeOut)
	if err != nil {
		logger.Errorf("解析超时时间失败:%v", err)
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), t)
	return ctx, cancel, nil
}

func sendHttpRequest(request *http.Request) *HttpResponse {
	logger.Infoln("开始发送请求......")
	resp, err := httpClient.Do(request)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		logger.Errorf("请求发送失败:%v", err)
		return &HttpResponse{500, nil, nil, err}
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf("body读取失败:%v", err)
			return &HttpResponse{resp.StatusCode, resp.Header, nil, err}
		}
		//log.Println("content-length:", resp.ContentLength)
		logger.Infoln("请求发送成功")
		return &HttpResponse{resp.StatusCode, resp.Header, body, nil}
	}
}

func DownloadFile(url string, destPath string) error {
	res, err := httpClient.Get(url)
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()
	if err != nil {
		return err
	}

	f, err := os.Create(destPath)

	defer func() {
		f.Close()
	}()

	if err != nil {
		return err
	}

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}
	return nil
}

func DealTime(t string) (time.Duration, error) {
	var (
		unit string
		r    int
		err  error
	)
	dealTimeErr := errors.New("数据格式不匹配")
	if len(t) <= 1 {
		return -1, dealTimeErr
	}
	if r, err = strconv.Atoi(t[:len(t)-1]); err != nil {
		r, _ = strconv.Atoi(t[:len(t)-2])
		unit = t[len(t)-2:]
	} else {
		unit = t[len(t)-1:]
	}
	switch unit {
	case "ms":
		return time.Duration(r) * time.Millisecond, nil
	case "s":
		return time.Duration(r) * time.Second, nil
	case "m":
		return time.Duration(r) * time.Minute, nil
	case "h":
		return time.Duration(r) * time.Hour, nil
	default:
		return -1, dealTimeErr
	}
}
