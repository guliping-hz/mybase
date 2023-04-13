package net2

import "sync"

type StatusNO int32

const (
	StatusUnknown  StatusNO = iota //0 尚未初始化
	StatusNormal                   //1 正常
	StatusShutdown                 //2 自己关闭的
	StatusTimeout                  //3 超时连接断开
	StatusError                    //4 其他异常断开,服务器主动断开 err为ErrClose
)

type Status struct {
	status StatusNO //StatusNormal
	err    error    //当为其他异常时，这里会有赋值
	sta    []byte   //调用的错误堆栈信息
	mutex  sync.RWMutex
}

func (s *Status) Error() string {
	if s.err == nil {
		return ""
	}
	return s.err.Error()
}

func (s *Status) Stack() []byte {
	return s.sta
}

func (s *Status) GetStatus() StatusNO {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.status
}

func (s *Status) ChangeStatusAll(status StatusNO, err error, stack []byte) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//只记录正常状态或者赋值为初始状态。
	if s.status == StatusNormal || s.status == StatusUnknown || status == StatusUnknown {
		s.status = status
		s.err = err
		s.sta = stack
		return true
	}
	return false
}

func (s *Status) ChangeStatus(status StatusNO, err error) bool {
	return s.ChangeStatusAll(status, err, nil)
}

func (s *Status) Reset() {
	s.ChangeStatusAll(StatusUnknown, nil, nil)
}
