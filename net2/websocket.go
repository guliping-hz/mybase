package net2

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type ClientWSocket struct {
	ClientBase
	msgType int //TextMessage or BinaryMessage
	conn    *websocket.Conn
}

// LocalAddr returns the local network address.
func (c *ClientWSocket) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *ClientWSocket) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ClientWSocket) Close() error {
	return c.conn.Close()
}

func (c *ClientWSocket) sendEx(buffer []byte) {
	//写超时必有
	err := c.conn.SetWriteDeadline(time.Now().Add(c.context.ttl))
	if err != nil {
		c.CloseWithErr(err, nil, true)
		return
	}

	err = c.conn.WriteMessage(c.msgType, buffer)
	//packageLen := binary.BigEndian.Uint16(buffer)
	//util.D("send buffer %d==%d,buf=%x", n, packageLen, buffer)
	if err != nil {
		c.CloseWithErr(err, nil, true)
		return
	}
}

func (c *ClientWSocket) recvEx() ([]byte, error) {
	if c.context.rTtl != 0 { //如果需要判断读超时。
		err := c.conn.SetReadDeadline(time.Now().Add(c.context.rTtl))
		if err != nil {
			return nil, err
		}
	}

	_, buffer, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// @msgType TextMessage or BinaryMessage
func (c *ClientWSocket) Connect(addr string, msgType int, ttl time.Duration, OnSocket OnSocket, ddb DataDecodeBase) error {
	if OnSocket == nil {
		return ErrParam
	}

	c.Init(ddb, ttl, 0, OnSocket, c, c)
	c.msgType = msgType
	return c.ReConnect(addr)
}

func (c *ClientWSocket) ReConnect(addr string) error {
	var err error

	c.conn, _, err = websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		//if err1, ok := err.(*net.OpError); ok && err1.Timeout() {
		//	return ErrTimeout
		//}
		return err
	}

	c.Reactor()
	return nil
}

// @msgType TextMessage or BinaryMessage
func WebAgent(conn *websocket.Conn, msgType int, ttl time.Duration, rTtl time.Duration, OnSocket OnSocket, ddb DataDecodeBase) *ClientWSocket {
	if OnSocket == nil {
		return nil
	}

	if ddb == nil {
		ddb = new(DataDecodeBinaryBigEnd)
	}
	csb := &ClientWSocket{
		msgType: msgType,
		conn:    conn,
	}
	csb.Init(ddb, ttl, rTtl, OnSocket, csb, csb)
	csb.Reactor()
	return csb
}
