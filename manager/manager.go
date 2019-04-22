package manager

import (
	logger "ChatServer/logger"
	"ChatServer/protobuf"
	"ChatServer/status"
	"sync"
	"time"
)

// Instance Is a singleton
type Instance struct {
	manager       map[chan<- *protobuf.Header]bool
	loginChan     chan chan<- *protobuf.Header
	logoutChan    chan chan<- *protobuf.Header
	broadcastChan chan *protobuf.Header
}

var once sync.Once
var manager *Instance

// InitManager init the manager instance
func InitManager() {
	once.Do(func() {
		manager = &Instance{}
		manager.init()
	})
}

// DoLogin add a channel to manager
func DoLogin(user chan<- *protobuf.Header) int32 {
	code := status.OK

	select {
	case manager.loginChan <- user:
	case <-time.After(time.Second * 10):
		logger.ErrorLog("Manager login timeout")
		code = status.TimeOut
	}
	return code
}

// DoLogout add a channel to logoutChan
func DoLogout(user chan<- *protobuf.Header) int32 {
	code := status.OK

	select {
	case manager.logoutChan <- user:
	case <-time.After(time.Second * 10):
		logger.ErrorLog("Manager timeout")
		code = status.TimeOut
	}
	return code
}

// DoBroadcast add a channel to logoutChan
func DoBroadcast(req *protobuf.Header) int32 {
	code := status.OK

	select {
	case manager.broadcastChan <- req:
	case <-time.After(time.Second * 10):
		logger.ErrorLog("Manager broadcast timeout")
		code = status.TimeOut
	}
	return code
}

func (instance *Instance) init() {
	instance.manager = make(map[chan<- *protobuf.Header]bool)
	instance.loginChan = make(chan chan<- *protobuf.Header)
	instance.logoutChan = make(chan chan<- *protobuf.Header)
	instance.broadcastChan = make(chan *protobuf.Header)
	go instance.worker()
}

func (instance *Instance) worker() {
	for {
		select {
		case user, ok := <-instance.loginChan:
			if ok == false {
				panic("Login channel should never be close")
			}
			instance.onLogin(user)
		case user, ok := <-instance.logoutChan:
			if ok == false {
				panic("Logout channel should never be close")
			}
			instance.onLogout(user)
		case req, ok := <-instance.broadcastChan:
			if ok == false {
				panic("Broadcast channel should never be close")
			}
			instance.onBroadcast(req)
		}
	}
}

func (instance *Instance) onLogin(user chan<- *protobuf.Header) int32 {
	TAG := "[Instance.onLogin] "
	instance.manager[user] = true
	logger.InfoLog(TAG, "Client login from manager")
	return status.OK
}

func (instance *Instance) onLogout(user chan<- *protobuf.Header) int32 {
	TAG := "[Instance.onLogout] "
	delete(instance.manager, user)
	logger.InfoLog(TAG, "Client logout from manager")
	return status.OK
}
func (instance *Instance) onBroadcast(req *protobuf.Header) int32 {
	for user := range instance.manager {
		select {
		case user <- req:
		default:
			close(user)
			delete(instance.manager, user)
			logger.ErrorLog("A dead channel is been close")
		}
	}
	return status.OK
}
