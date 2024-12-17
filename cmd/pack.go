package cmd

import (
	"encoding/binary"
	"github.com/guliping-hz/mybase"
	"github.com/guliping-hz/mybase/net2"
	"google.golang.org/protobuf/proto"
	"reflect"
)

func PackSeqData(cmd interface{}, seq int32, data proto.Message) *PackMsg {
	bs, _ := proto.Marshal(data)
	return &PackMsg{
		Cmd:    int32(reflect.ValueOf(cmd).Int()),
		Seq:    seq,
		Binary: bs,
	}
}

func PackData(cmd interface{}, data proto.Message) *PackMsg {
	bs, _ := proto.Marshal(data)
	return &PackMsg{
		Cmd:    int32(reflect.ValueOf(cmd).Int()),
		Binary: bs,
	}
}

func PackSeqPackage(id interface{}, seq int32, data proto.Message) []byte {
	content := PackSeqData(id, seq, data)
	return PackProtoToPackage(content, int32(reflect.ValueOf(id).Int()))
}

func PackPackage(id interface{}, data proto.Message) []byte {
	content := PackData(id, data)
	return PackProtoToPackage(content, int32(reflect.ValueOf(id).Int()))
}

func PackContentToPackage(bufContent []byte) []byte {
	contentLen := len(bufContent)
	if contentLen > 65533 { //一个包最大长度不能超过65535-2
		return nil
	}
	bufPackage := make([]byte, net2.GetDefaultPackageHeadLen()+contentLen)
	binary.BigEndian.PutUint16(bufPackage, uint16(contentLen))
	copy(bufPackage[2:], bufContent)

	//fmt.Printf("%v\n", bufPackage)
	return bufPackage //返回成功消息
}

func PackProtoToPackage(message proto.Message, cmdForLog int32) []byte {
	buf, err := proto.Marshal(message)
	if err != nil {
		mybase.E("Marshal fail err=%s", err.Error())
		return nil
	}

	buf = PackContentToPackage(buf)
	if buf == nil { //一个包最大长度不能超过65535-2
		mybase.E("the buf is over 65535 cmd=%d", cmdForLog)
		return nil
	}
	return buf
}

func SendToClient(conn net2.Conn, message *PackMsg, cmdForLog int32) bool {
	buf := PackProtoToPackage(message, cmdForLog)
	if buf == nil {
		return false
	}

	//mybase.I("send single buf=%d", len(buf))
	//if time.Now().Unix() < attr.InsDebugUtc {
	//	mybase.T("SendToClient %+v cmd:%d", message, cmdForLog)
	//}
	//
	//if attr.InsNeedSleep {
	//	time.Sleep(time.Millisecond)
	//}

	ok := conn.Send(buf)
	if !ok {
		return false
	}
	return true
}

func SendToSingleServer(conn net2.Conn, message proto.Message) bool {
	buf := PackProtoToPackage(message, 0)
	if buf == nil {
		return false
	}

	//mybase.I("send single buf=%d", len(buf))
	//if time.Now().Unix() < attr.InsDebugUtc {
	//	mybase.T("SendToSingleServer %+v cmd:%d", message, 0)
	//}
	//
	//if attr.InsNeedSleep {
	//	time.Sleep(time.Millisecond)
	//}

	ok := conn.Send(buf)
	if !ok {
		return false
	}
	return true
}

func SendToSingleServerNoLen(conn net2.Conn, message *AgentData) bool {
	buf, err := proto.Marshal(message)
	if err != nil {
		mybase.E("Marshal fail err=%s", err.Error())
		return false
	}

	//if time.Now().Unix() < attr.InsDebugUtc {
	//	mybase.T("SendToSingleServerNoLen %+v", message)
	//}
	//
	//if attr.InsNeedSleep {
	//	time.Sleep(time.Millisecond)
	//}

	ok := conn.Send(buf)
	if !ok {
		return false
	}
	return true
}
