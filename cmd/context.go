package cmd

import (
	"fmt"
	"github.com/guliping-hz/mybase/net2"
	"google.golang.org/protobuf/proto"
	"math"
	"reflect"
)

type Handler func() bool
type HandlersChan []Handler

const abortIndex int8 = math.MaxInt8 / 2

type Context struct {
	net2.Conn
	Head *PackMsg

	handlers HandlersChan
	index    int8

	sessionIdStr string
}

func (c *Context) Next() bool {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if ok := c.handlers[c.index](); !ok {
			return true
		}
		c.index++
	}
	return true
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) ResetNil() {
	c.Head = nil
	c.handlers = nil
	c.index = -1
}

func (c *Context) Reset(msg *PackMsg, handlers HandlersChan) error {
	if len(handlers) >= int(abortIndex) {
		return fmt.Errorf("handlers len is over %d", abortIndex)
	}

	c.ResetNil()

	c.Head = msg
	c.handlers = handlers
	return nil
}

func (c *Context) resetPackMsg() {
	if c.Head == nil {
		c.Head = new(PackMsg)
	} else {
		c.Head.Cmd = 0
		c.Head.Seq = 0
		c.Head.Binary = nil
		c.Head.Ret = 0
		c.Head.Tip = ""
	}
}

func (c *Context) BackRet(ret int32) bool {
	c.Head.Ret = ret
	c.Head.Binary = nil
	return c.SendPackMsg(c.Head)
}

//func (c *Context) BackTip(tip string) bool {
//	return c.BackRetTip(db.ErrorTip, tip)
//}

func (c *Context) BackRetTip(ret int32, tip string) bool {
	c.Head.Ret = ret
	c.Head.Tip = tip
	c.Head.Binary = nil
	return c.SendPackMsg(c.Head)
}

func (c *Context) BackData(data proto.Message) bool {
	bs, _ := proto.Marshal(data)

	return c.BackDataBuf(bs)
}

func (c *Context) BackDataBuf(data []byte) bool {
	c.Head.Ret = 0
	c.Head.Binary = data
	return c.SendPackMsg(c.Head)
}

// 去掉客户端的seq倒计时，但是没有对应的CMD回调
func (c *Context) BackWait() bool {
	c.Head.Cmd = 0
	c.Head.Ret = 0
	return c.SendPackMsg(c.Head)
}

// 当前不回应客户端，保持代码完整性
func (c *Context) BackWait0() bool {
	c.Head.Cmd = 0
	c.Head.Ret = 0
	c.Head.Seq = 0
	return c.SendPackMsg(c.Head)
}

func (c *Context) SendRet(cmd interface{}, ret int32) bool {
	return c.SendPackMsg(&PackMsg{Cmd: int32(reflect.ValueOf(cmd).Int()), Ret: ret})
}

func (c *Context) SendTip(cmd interface{}, ret int32, tip string) bool {
	return c.SendPackMsg(&PackMsg{Cmd: int32(reflect.ValueOf(cmd).Int()), Ret: ret, Tip: tip})
}

func (c *Context) SendData(cmd interface{}, data proto.Message) bool {
	return c.SendPackMsg(PackData(cmd, data))
}

func (c *Context) SendPackMsg(head *PackMsg) bool {
	c.Abort() //中止

	SendToClient(c.Conn, head, head.Cmd)

	return true //返回成功消息
}

// 线程安全版本
func (c *Context) SendBuff(bufPackage []byte) bool { //中止
	//if c.Uid == 120131 || c.Uid == 120132 {
	//fmt.Printf("send buff[%d]=%v\n", len(bufPackage), bufPackage)
	//}
	//if time.Now().Unix() < attr.InsDebugUtc {
	//	mybase.T("send buff:0x%x", bufPackage)
	//}
	//
	//if attr.InsNeedSleep {
	//	time.Sleep(time.Millisecond)
	//}

	return c.Send(bufPackage) //返回成功消息
}
