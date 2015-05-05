package punchy

import (
	"bytes"
	"encoding/gob"
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
