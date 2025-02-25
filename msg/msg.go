package msg

import (
	"encoding/binary"
	"io"
)

type messageID  uint8

const (
	MSG_CHOKE messageID = iota
	MSG_UNCHOKE
	MSG_INTERESTED
	MSG_NOT_INTERESTED
	MSG_HAVE
	MSG_BITFIELD
	MSG_REQUEST
	MSG_PIECE
	MSG_CANCEL

)

// Message stores ID and payload of a message
type Message struct {
    ID      messageID

	//payload of the message. may be optional
    Payload []byte 
}

func (m *Message) Sereialize()  []byte {

	if m == nil {
        return make([]byte, 4)
    }
	length := uint32(len(m.Payload) + 1) // +1 for id

	buff := make([]byte, 4+length) //4 bytes (32bits) to rep length 
    binary.BigEndian.PutUint32(buff[0:4], length)

	buff[4] = byte(m.ID)
	copy(buff[5:], m.Payload)

	return buff
}

func (m *Message) DeSerialize(r io.Reader) error {

	lengthBuf := make([]byte, 4)
    _, err := io.ReadFull(r, lengthBuf)
    if err != nil {
        return err
    }
    length := binary.BigEndian.Uint32(lengthBuf)

    // keep-alive message
    if length == 0 {
		m = nil
        return nil
    }

    messageBuf := make([]byte, length)
    _, err = io.ReadFull(r, messageBuf)
    if err != nil {
        return err
    }

	m.ID = messageID(messageBuf[0])
	m.Payload = messageBuf[1:]

    return nil
}