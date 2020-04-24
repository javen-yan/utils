package consul

import (
	"context"
	"github.com/yanzongzhen/Logger/logger"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct{}

func (h *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	logger.Infof(req.Service)
	return nil
}

//func (c *Consul) InitHealthCheck() {
//	logger.Debugf("初始化健康检查端口:%v", c.Options.HealthPort)
//	go func() {
//		l, err := net.Listen("tcp", ":"+strconv.Itoa(c.Options.HealthPort))
//		if err != nil {
//			logger.Errorf("开启健康检查服务失败:%v", err)
//			return
//		}
//		s := grpc.NewServer()
//		grpc_health_v1.RegisterHealthServer(s, &healthServer{})
//		err = s.Serve(l)
//		if err != nil {
//			logger.Errorf("健康检查初始化失败:%v", err)
//		}
//	}()
//}
