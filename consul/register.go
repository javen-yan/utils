package consul

import (
	"errors"
	"fmt"
	"github.com/yanzongzhen/Logger/logger"
	"github.com/hashicorp/consul/api"
	"net"
	"strconv"
)

type Consul struct {
	Client  *api.Client
	Options Options
}
type Options struct {
	ServicePort  int
	HealthPort   int
	HealthMethod string
	ServiceName  string
	Type         string
}

func NewConsul(Options Options, config ...api.Config) (c *Consul, err error) {
	var Client *api.Client
	if len(config) == 0 {
		Client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return
		}
	} else {
		Client, err = api.NewClient(&config[0])
		if err != nil {
			return
		}
	}
	return &Consul{
		Client:  Client,
		Options: Options,
	}, nil
}

func (c *Consul) ServiceRegister() (string, error) {
	logger.Debugf("服务注册:%v", c.Options.ServiceName)
	//ip
	ip, err := getLocalAddr()
	if err != nil {
		return "", err
	}
	//serviceId
	serviceID := c.Options.ServiceName + "-" + ip + ":" + strconv.Itoa(c.Options.ServicePort) + "-" + c.Options.Type
	logger.Debugf("serviceId:%v,name:%v,addr:%v", serviceID, c.Options.ServiceName, ip+":"+strconv.Itoa(c.Options.ServicePort))
	//创建一个新服务。
	registration := new(api.AgentServiceRegistration)
	registration.ID = serviceID
	registration.Name = c.Options.ServiceName
	registration.Port = c.Options.ServicePort
	registration.Tags = nil
	registration.Address = ip
	//增加check。
	check := new(api.AgentServiceCheck)
	if c.Options.Type == "rpc" {
		check.GRPC = fmt.Sprintf("%s:%d%s", registration.Address, c.Options.HealthPort, c.Options.HealthMethod)
	} else if c.Options.Type == "http" {
		check.HTTP = fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, c.Options.HealthMethod)
	} else {
		return "", errors.New("service type not favor")
	}
	//设置超时 5s。
	check.Timeout = "5s"
	//设置间隔 5s。
	check.Interval = "10s"
	//注册check服务。
	registration.Check = check

	err = c.Client.Agent().ServiceRegister(registration)

	if err != nil {
		logger.Errorf("服务注册失败:%v", err)
		return "", err
	}
	return serviceID, nil
}

func (c *Consul) ServiceUnRegister() (err error) {
	logger.Debugf("解除服务注册:%v", c.Options.ServiceName)
	ip, err := getLocalAddr()
	if err != nil {
		return err
	}
	err = c.Client.Agent().ServiceDeregister(c.Options.ServiceName + "-" + ip + ":" + strconv.Itoa(c.Options.ServicePort) + "-" + c.Options.Type)
	return
}

func getLocalAddr() (string, error) {
	var ip string
	addrSlice, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorf("获取本地IP地址失败:%v", err)
		return "", err
	}
	for _, addr := range addrSlice {
		if iPNet, ok := addr.(*net.IPNet); ok && !iPNet.IP.IsLoopback() {
			if nil != iPNet.IP.To4() {
				ip = iPNet.IP.String()
				logger.Debugf("本机IP为:%v", ip)
				return ip, nil
			}
		}
	}
	return "", errors.New("获取本地IP地址失败")
}
