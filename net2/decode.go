package net2

import "encoding/binary"

//文本解析 => 没有长度的概念+包内容
type DataDecodeText struct {
}

func (d *DataDecodeText) GetPackageHeadLen() int {
	return 0
}
func (d *DataDecodeText) GetPackageLen(buf []byte) int {
	return len(buf)
}

//二进制大端包 => 2字节包长+包内容
type DataDecodeBinaryBigEnd struct {
}

func (d *DataDecodeBinaryBigEnd) GetPackageHeadLen() int {
	return GetDefaultPackageHeadLen()
}
func (d *DataDecodeBinaryBigEnd) GetPackageLen(buf []byte) int {
	//return len(buf)
	return int(binary.BigEndian.Uint16(buf))
}

func GetDefaultPackageHeadLen() int {
	return 2 //2字节表示包长 一个包最大长度为：65535
}
