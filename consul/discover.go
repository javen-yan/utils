package consul

import (
	"github.com/yanzongzhen/Logger/logger"
	"github.com/hashicorp/consul/api"
	"strconv"
	"strings"
	"time"
)

type ServiceInfo struct {
	Addr        string
	ServiceId   string
	ServiceName string
}

func (c *Consul) ServiceDiscover(prefix string) ([]*ServiceInfo, error) {
	servicesInfo := make([]*ServiceInfo, 0)
	services, _, err := c.Client.Catalog().Services(&api.QueryOptions{})
	if err != nil {
		return nil, err
	}
	for name := range services {
		servicesData, _, err := c.Client.Health().Service(name, "", true,
			&api.QueryOptions{})
		if err != nil {
			return nil, err
		}
		for _, entry := range servicesData {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			for _, health := range entry.Checks {
				if health.ServiceName != name {
					continue
				}
				addr := entry.Service.Address + ":" + strconv.Itoa(entry.Service.Port)
				serviceName := health.ServiceName
				serviceId := health.ServiceID
				servicesInfo = append(servicesInfo, &ServiceInfo{
					Addr:        addr,
					ServiceName: serviceName,
					ServiceId:   serviceId,
				})
			}
		}
	}
	for _, s := range servicesInfo {
		logger.Debugf("服务发现:%v", s.Addr)
	}
	return servicesInfo, nil
}

func DiscoverService(c *Consul, waitForDiscover bool, prefix string) (map[string]string, error) {
	var services []*ServiceInfo
	var err error
	for {
		services, err = c.ServiceDiscover(prefix)
		if err != nil {
			logger.Errorf("服务发现失败:%v", err)
			return nil, err
		}
		if waitForDiscover {
			if len(services) != 0 {
				break
			}
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	binds := make(map[string]string)
	if len(services) > 0 {
		for _, s := range services {
			binds[s.ServiceId] = s.Addr
		}
	}
	return binds, nil
}
