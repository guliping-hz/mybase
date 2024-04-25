package net2

import (
	"log"
	"sync"
	"testing"
	"time"
)

type TestCli struct {
	ClientSocket
}

func (t *TestCli) OnConnect(conn Conn) {
	log.Printf("OnConnect %s\n", conn.UnionId())
}

func (t *TestCli) OnClose(conn Conn, byLocalNotRemote bool) {
	log.Printf("OnClose %s byLocalNotRemote:%v\n", conn.UnionId(), byLocalNotRemote)

	if !byLocalNotRemote {
		t.ConnectLobby()
	}
}

func (t *TestCli) OnTimeout(conn Conn) {
	log.Printf("OnTimeout %s\n", conn.UnionId())
}

func (t *TestCli) OnNetErr(conn Conn) {
	log.Printf("OnNetErr %s\n", conn.UnionId())
}

func (t *TestCli) OnRecvMsg(conn Conn, buf []byte) bool {
	log.Printf("OnRecvMsg %s buf:%x\n", conn.UnionId(), buf)
	return true
}

func (t *TestCli) ConnectLobby() {
	log.Printf("try connect\n")
	if err := t.Connect("127.0.0.1:5102", time.Second*60, t, nil); err != nil {
		go func() {
			time.Sleep(time.Second * 3)
			t.ConnectLobby()
		}()
	}
}

func TestClientSocket_ReConnect(t *testing.T) {
	wait := sync.WaitGroup{}
	wait.Add(1)
	cli := new(TestCli)
	cli.ConnectLobby()
	wait.Wait()
}
