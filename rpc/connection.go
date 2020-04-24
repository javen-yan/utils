package rpc

import (
	"errors"
	"github.com/yanzongzhen/Logger/logger"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var BadServerError = errors.New("bad server")

const pingInterval = 20 * time.Second

type PingFunc func(conn *grpc.ClientConn) error
type RequestFunc func(conn *grpc.ClientConn) error

var mapLock sync.RWMutex

var clientMap map[string]*rpcConn

func init() {
	clientMap = make(map[string]*rpcConn)
}

type rpcConn struct {
	conn     *grpc.ClientConn
	server   string
	error    error
	stop     chan int
	diaLock  *sync.Mutex
	isDelete bool
}

func (conn *rpcConn) dial() error {
	conn.diaLock.Lock()
	defer conn.diaLock.Unlock()

	if conn.error == nil {
		return nil
	}

	rpcConn, err := grpc.Dial(conn.server, grpc.WithInsecure())
	if err != nil {
		conn.error = err
		return err
	}
	conn.conn = rpcConn
	conn.error = nil
	return nil
}

func (conn *rpcConn) ping(pingFunc PingFunc) {
	for {
		select {
		case <-conn.stop:
			return
		default:
			break
		}

		if conn.error == nil {
			conn.error = pingFunc(conn.conn)
		} else {
			if conn.conn != nil {
				_ = conn.conn.Close()
			}
			conn.error = conn.dial()
		}
		time.Sleep(pingInterval)
	}
}

func (conn *rpcConn) disConnect() {
	close(conn.stop)
	if conn.conn != nil {
		_ = conn.conn.Close()
	}
}

func InitConnection(ipList []string, pingFunc PingFunc) error {
	return InitConnectionWithTag(ipList, nil, pingFunc)
}

func InitConnectionWithTag(ipList []string, tag []string, pingFunc PingFunc) error {
	if tag != nil && len(ipList) != len(tag) {
		return errors.New("not match error")
	}
	for i, host := range ipList {
		c := newConnection(host)
		err := c.dial()
		if err != nil {
			return err
		}
		if tag != nil {
			clientMap[host+tag[i]] = c
			go c.ping(pingFunc)
		} else {
			clientMap[host] = c
			go c.ping(pingFunc)
		}
	}
	return nil
}

func newConnection(url string) *rpcConn {
	c := &rpcConn{}
	c.error = errors.New("first")
	c.server = url
	c.diaLock = &sync.Mutex{}
	c.stop = make(chan int)
	c.isDelete = false
	return c
}

func getRpcConnectionWithTag(url string, tag string) *grpc.ClientConn {

	mapLock.RLock()
	client, ok := clientMap[url+tag]
	logger.Debugln(clientMap)
	mapLock.RUnlock()
	if ok {
		if client.error == nil {
			return client.conn
		} else {
			if client.conn != nil {
				_ = client.conn.Close()
			}
			err := client.dial()
			if err != nil {
				return nil
			}
			return client.conn
		}
	}
	return nil
}

func getRpcConnection(url string) *grpc.ClientConn {
	return getRpcConnectionWithTag(url, "")
}

func DoRpcRequestWithTag(url string, tag string, requestFunc RequestFunc) error {
	conn := getRpcConnectionWithTag(url, tag)
	if conn != nil {
		err := requestFunc(conn)
		if err != nil {
			return err
		}
		return nil
	} else {
		return BadServerError
	}
}

func DoRpcRequest(url string, requestFunc RequestFunc) error {
	return DoRpcRequestWithTag(url, "", requestFunc)
}

func UpdateRpcConnectionWithTag(ipList []string, tag []string, pingFunc PingFunc) error {
	mapLock.Lock()
	defer mapLock.Unlock()

	if len(ipList) != len(tag) {
		return errors.New("not match error")
	}

	for _, c := range clientMap {
		c.isDelete = true
	}

	isTag := tag != nil

	for i, host := range ipList {
		var c *rpcConn
		var ok bool
		if isTag {
			c, ok = clientMap[host+tag[i]]
		} else {
			c, ok = clientMap[host]
		}
		if !ok {
			c = newConnection(host)
			go c.ping(pingFunc)
			if isTag {
				clientMap[host] = c
			} else {
				clientMap[host+tag[i]] = c
			}
		} else {
			c.isDelete = false
		}
	}

	for k, c := range clientMap {
		if c.isDelete {
			c.disConnect()
			delete(clientMap, k)
		}
	}
	return nil
}

func UpdateRpcConnection(servers []string, pingFunc PingFunc) error {
	mapLock.Lock()
	defer mapLock.Unlock()

	for _, c := range clientMap {
		c.isDelete = true
	}

	for _, host := range servers {
		logger.Debug(host)
		c, ok := clientMap[host]
		logger.Debug(ok)
		if !ok {
			c = newConnection(host)
			go c.ping(pingFunc)
			clientMap[host] = c
		} else {
			c.isDelete = false
		}
	}
	logger.Debug(clientMap)

	for k, c := range clientMap {
		logger.Debug(c.isDelete)
		if c.isDelete {
			c.disConnect()
			delete(clientMap, k)
		}
	}
	return nil
}
