package net2

import (
	"fmt"
	"net"
	"time"
)

func CheckTimeout(err error) bool {
	if err != nil {
		if err1, ok := err.(*net.OpError); ok {
			return err1.Timeout()
		}
	}
	return false
}

// *********ClientSocket
type ClientSocket struct {
	ClientBase
	conn net.Conn
}

func (c *ClientSocket) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
func (c *ClientSocket) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *ClientSocket) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ClientSocket) sendEx(buffer []byte) {
	//写超时必有
	err := c.conn.SetWriteDeadline(time.Now().Add(c.context.ttl))
	if err != nil {
		c.CloseWithErr(err, nil, true)
		return
	}

	_, err = c.conn.Write(buffer)
	//fmt.Printf("SendEx buf[%d]\n", n)
	if err != nil {
		if CheckTimeout(err) {
			c.CloseTimeout()
		} else {
			c.CloseWithErr(err, nil, true)
		}
		return
	}
}

func (c *ClientSocket) recvEx() ([]byte, error) {
	if c.context.rTtl != 0 { //如果需要判断读超时。
		err := c.conn.SetReadDeadline(time.Now().Add(c.context.rTtl))
		if err != nil {
			return nil, err
		}
	}

	buffer := make([]byte, 2048)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:n], nil
}

func (c *ClientSocket) ConnectHostPort(host string, port uint16, Ttl time.Duration, OnSocket OnSocket, ddb DataDecodeBase) error {
	//通过域名找IP地址
	ip, err := net.ResolveIPAddr("", host)
	if err != nil {
		return err
	}
	var addr = fmt.Sprintf("%s:%d", ip.IP.String(), port)

	return c.Connect(addr, Ttl, OnSocket, ddb)
}

func (c *ClientSocket) Connect(addr string, ttl time.Duration, OnSocket OnSocket, ddb DataDecodeBase) error {
	if OnSocket == nil {
		panic("OnSocket is nil")
	}

	c.Init(ddb, ttl, 0, OnSocket, c, c)
	return c.ReConnect(addr)
}

func (c *ClientSocket) ReConnect(addr string) error {
	var err error
	c.conn, err = net.DialTimeout("tcp", addr, c.context.ttl)
	if err != nil {
		if CheckTimeout(err) {
			return ErrTimeout
		}
		return err
	}

	c.Reactor()
	return nil
}

func Agent(conn net.Conn, ttl time.Duration, rTtl time.Duration, OnSocket OnSocket, ddb DataDecodeBase) *ClientSocket {
	if OnSocket == nil {
		return nil
	}

	if ddb == nil {
		ddb = new(DataDecodeBinaryBigEnd)
	}

	csb := &ClientSocket{
		conn: conn,
	}
	csb.Init(ddb, ttl, rTtl, OnSocket, csb, csb)
	return csb
}
