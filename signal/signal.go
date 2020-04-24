package signal

import (
	"github.com/yanzongzhen/Logger/logger"
	"os"
	"os/signal"
	"syscall"
)

var signalFuncList []func()

func init() {
	signalFuncList = make([]func(), 0, 10)
}

func AddSignalFunc(signalFunc func()) {
	signalFuncList = append(signalFuncList, signalFunc)
}

func WaitSignal() {
	logger.Debug("wait signal")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP, syscall.SIGKILL)
	for {
		select {
		case a := <-c:
			logger.Debugf("接受到退出信号:%v", a.String())
			logger.Debug(len(signalFuncList))
			for _, s := range signalFuncList {
				logger.Debug("run")
				s()
			}
			return
		}
	}
}
