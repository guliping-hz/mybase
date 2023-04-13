package net2

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type StatusConnServer struct {
	listener *net.TCPListener
	Status
}

type StackError interface {
	error
	Stack() []byte
}

type OnSocketServer interface {
	OnSocket
	OnServerListen()
	OnServerErr(StackError)
	//这里OnServerClose始终会回调
	OnServerClose()
}

type ServerSocket struct {
	StatusConnServer
	onSocket OnSocketServer
	ttl      time.Duration //监听客户端读取超时时间，如果不需要有超时机制，可以设置为0
	rTtl     time.Duration

	chanAccept chan net.Conn
	chanStop   chan bool

	fd2Client sync.Map //int64 net2.Context.SessionId() -> true

	clientDataDecoder DataDecodeBase
	listenAddress     string
}

/*
连接上服务器回调
*/
func (s *ServerSocket) OnConnect(conn Conn) {
	s.fd2Client.Store(conn.SessionId(), conn)
	s.onSocket.OnConnect(conn)
}

/*
只要我们曾经连接上服务器过，OnClose必定会回调。代表一个当前的socket已经关闭
*/
func (s *ServerSocket) OnClose(conn Conn, byLocalNotRemote bool) {
	s.fd2Client.Delete(conn.SessionId())
	s.onSocket.OnClose(conn, byLocalNotRemote)
}

/*
连接超时,写入超时,读取超时回调，之后会调用OnClose
*/
func (s *ServerSocket) OnTimeout(conn Conn) {
	s.onSocket.OnTimeout(conn)
}

/*
网络错误回调，之后会调用OnClose
*/
func (s *ServerSocket) OnNetErr(conn Conn) {
	s.onSocket.OnNetErr(conn)
}

/*
接受到信息
@return 返回true表示可以继续热恋，false表示要分手了。
*/
func (s *ServerSocket) OnRecvMsg(conn Conn, buf []byte) bool {
	return s.onSocket.OnRecvMsg(conn, buf)
}

func (s *ServerSocket) Shutdown() {
	//停止服务器运行。
	s.chanStop <- true
}

func (s *ServerSocket) Listen() error {
	if s.onSocket == nil {
		return ErrParam
	}
	//获取服务器监听ip地址
	ip, err := net.ResolveTCPAddr("", s.listenAddress)
	if err != nil {
		return err
	}

	//创建一个监听的socket
	s.listener, err = net.ListenTCP("tcp", ip)
	if err != nil {
		return err
	}

	go s.reactor()
	return nil
}

func (s *ServerSocket) close() {
	err := s.listener.Close()
	if err != nil {
		log.Printf("Close error=%v\n", err.Error())
	}
	s.onSocket.OnServerClose()
}

func (s *ServerSocket) reactor() {
	defer s.close()

	s.onSocket.OnServerListen()
	go func() {
		for {
			clientConn, err := s.listener.Accept()
			if err != nil {
				s.ChangeStatus(StatusError, err)
				s.onSocket.OnServerErr(s)
				return
			}
			s.chanAccept <- clientConn
		}
	}()

loop:
	for {
		select {
		case clientConn := <-s.chanAccept:
			//@todo agent可以考虑弄个代理池，或许更高效一点
			agent := Agent(clientConn, s.ttl, s.rTtl, s, s.clientDataDecoder)
			if agent != nil {
				agent.Reactor()
			}
		case <-s.chanStop:
			//关闭已经连接的
			s.fd2Client.Range(func(key, value interface{}) bool {
				agent := value.(*Context)
				agent.Con.SafeClose(false)
				return true
			})
			//关闭监听的socket
			_ = s.listener.Close()
			break loop
		}
	}
}

func NewServerIp(ip string, port uint16, onSocket OnSocketServer, clientDDB DataDecodeBase) *ServerSocket {
	return NewServer(fmt.Sprintf("%s:%d", ip, port), time.Second*30, time.Second*30, onSocket, clientDDB)
}

/*
@ttl 客户端发送超时
@rTtl 客户端读取超时 0表示永远等待读取。
*/
func NewServer(address string, ttl time.Duration, rTtl time.Duration, onSocket OnSocketServer,
	clientDDB DataDecodeBase) *ServerSocket {
	ssb := &ServerSocket{}
	ssb.listenAddress = address
	ssb.onSocket = onSocket
	ssb.chanStop = make(chan bool)
	ssb.chanAccept = make(chan net.Conn)
	ssb.clientDataDecoder = clientDDB
	ssb.ttl = ttl
	ssb.rTtl = rTtl
	return ssb
}
