/**
 * @Author: guanyunlong
 * @Description:
 * @File:  middleware
 * @Version: 1.0.0
 * @Date: 20-3-4 下午3:35
 */
package server

import (
	"fmt"
	"github.com/yanzongzhen/Logger/logger"
	"time"
)

type Interceptor func(ctx *Context)

//func Auth(ctx *Context) error {
//	if !config.IsNotCheckJWT(ctx.Request.URL.Path) {
//		token, err := jrequest.ParseFromRequest(ctx.Request, jrequest.AuthorizationHeaderExtractor,
//			func(token *jwt.Token) (interface{}, error) {
//				return []byte(config.JWTSecret), nil
//			})
//		if err != nil {
//			return err
//		}
//		if !token.Valid {
//			return errors.New("token auth failed")
//		}
//		claims := token.Claims.(jwt.MapClaims)
//		account, _ := claims["aud"].(string)
//		orgID, _ := claims["org"].(string)
//		deviceID, _ := claims["did"].(string)
//		deviceType, _ := claims["dty"].(string)
//		ctx.Params = append(ctx.Params, httprouter.Param{Key: "account", Value: account})
//		ctx.Params = append(ctx.Params, httprouter.Param{Key: "orgID", Value: orgID})
//		ctx.Params = append(ctx.Params, httprouter.Param{Key: "deviceID", Value: deviceID})
//		ctx.Params = append(ctx.Params, httprouter.Param{Key: "deviceType", Value: deviceType})
//	}
//	return nil
//}

func Time(server func(ctx *Context)) func(ctx *Context) {
	return func(ctx *Context) {
		startTime := time.Now().UnixNano()
		logger.Info(fmt.Sprintf("请求开始--请求地址为:%s", ctx.Request.URL.String()))
		server(ctx)
		logger.Info(fmt.Sprintf("请求结束--请求耗时为:%d ms", (time.Now().UnixNano()-startTime)/int64(time.Millisecond)))
	}
}
