package message

import (
	"encoding/binary"
	"fmt"
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

func (m *Message) Serialize()  []byte {

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

// parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}



// FormatRequest creates a REQUEST message
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{ID: MSG_REQUEST, Payload: payload}
}

// FormatHave creates a HAVE message
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MSG_HAVE, Payload: payload}
}

// ParsePiece parses a PIECE message and copies its payload into a buffer
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {

	//Piece: <index><begin><data>
	if msg.ID != MSG_PIECE {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MSG_PIECE, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

func ParseHave(m *Message) (int, error) {
	if m.ID != MSG_HAVE {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MSG_HAVE, m.ID)
	}
	if len(m.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(m.Payload))
	}

	index := int(binary.BigEndian.Uint32(m.Payload))
	return index, nil
}
                                  