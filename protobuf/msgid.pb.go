// Code generated by protoc-gen-go. DO NOT EDIT.
// source: msgid.proto

package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type MsgId int32

const (
	MsgId_NO_USE MsgId = 0
	// 通用消息id
	MsgId_KEEP_ALIVE   MsgId = 1
	MsgId_CS_LOGIN_REQ MsgId = 1001
	MsgId_SC_LOGIN_RSP MsgId = 2001
)

var MsgId_name = map[int32]string{
	0:    "NO_USE",
	1:    "KEEP_ALIVE",
	1001: "CS_LOGIN_REQ",
	2001: "SC_LOGIN_RSP",
}
var MsgId_value = map[string]int32{
	"NO_USE":       0,
	"KEEP_ALIVE":   1,
	"CS_LOGIN_REQ": 1001,
	"SC_LOGIN_RSP": 2001,
}

func (x MsgId) String() string {
	return proto.EnumName(MsgId_name, int32(x))
}
func (MsgId) EnumDescriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func init() {
	proto.RegisterEnum("protocol.MsgId", MsgId_name, MsgId_value)
}

func init() { proto.RegisterFile("msgid.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 116 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xce, 0x2d, 0x4e, 0xcf,
	0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x5a, 0x9e,
	0x5c, 0xac, 0xbe, 0xc5, 0xe9, 0x9e, 0x29, 0x42, 0x5c, 0x5c, 0x6c, 0x7e, 0xfe, 0xf1, 0xa1, 0xc1,
	0xae, 0x02, 0x0c, 0x42, 0x7c, 0x5c, 0x5c, 0xde, 0xae, 0xae, 0x01, 0xf1, 0x8e, 0x3e, 0x9e, 0x61,
	0xae, 0x02, 0x8c, 0x42, 0x82, 0x5c, 0x3c, 0xce, 0xc1, 0xf1, 0x3e, 0xfe, 0xee, 0x9e, 0x7e, 0xf1,
	0x41, 0xae, 0x81, 0x02, 0x2f, 0xd9, 0x41, 0x42, 0xc1, 0xce, 0x30, 0xa1, 0xe0, 0x00, 0x81, 0x8b,
	0xfc, 0x49, 0x6c, 0x60, 0x43, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8f, 0xef, 0x95, 0x35,
	0x6a, 0x00, 0x00, 0x00,
}
