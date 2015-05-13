package punchy

import (
	"bytes"
	"encoding/gob"
	"net"
)

type OutboundMessage interface {
	RawMessage() RawMessage
}

func (m *RoomMessage) RawMessage() (RawMessage, error) {
	w := new(bytes.Buffer)
	var raw RawMessage
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m.Room)
	if err != nil {
		return raw, err
	}
	raw.Data = w.Bytes()
	return raw, nil
}

type RoomListMessage struct {
	RoomMessage
	Length    uint16
	Addresses []net.UDPAddr
}

func (m *RoomListMessage) RawMessage() (RawMessage, error) {
	var raw RawMessage
	data, err := m.EncodeMessage()
	raw.Data = data
	return raw, err
}
func (m *RoomListMessage) EncodeMessage() ([]byte, error) {
	w := new(bytes.Buffer)
	enc := gob.NewEncoder(w)
	err := enc.Encode(m.Length)
	if err != nil {
		panic(err)
	}
	err = enc.Encode(m.Room)
	if err != nil {
		panic(err)
	}
	for i := uint16(0); i < m.Length; i++ {
		err = enc.Encode(&m.Addresses[i])
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}
	return w.Bytes(), nil
}

func (m *RoomListMessage) DecodeMessage(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&m.Length)
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(&m.Room)
	if err != nil {
		panic(err)
	}
	m.Addresses = make([]net.UDPAddr, m.Length)
	for i := uint16(0); i < m.Length; i++ {
		err = decoder.Decode(&m.Addresses[i])
		if err != nil {
			panic(err)
		}
	}

	return nil
}
