package punchy

import (
	"net"
)

type InboundMessage interface {
	Sender() *net.UDPAddr
	DataAsString() string
	RawData() []byte
	Size() int
	Encrypted() bool
	Type() MessageType
	DecodeMessage(*net.UDPAddr, []byte) error
}

func (m *Message) Sender() *net.UDPAddr {
	return m.RawMessage.Sender
}

func (m *Message) RawData() []byte {
	return m.Data
}

func (m *Message) DataAsString() string {
	return string(m.Data)
}

func (m *Message) Size() int {
	return len(m.Data)
}

func (m *Message) Encrypted() bool {
	return m.EncryptedMsg
}

func (m *Message) Type() MessageType {
	return m.MsgType
}
