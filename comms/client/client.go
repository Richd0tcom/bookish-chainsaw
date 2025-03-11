package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/Richd0tcom/bookish-chainsaw/bitfield"
	"github.com/Richd0tcom/bookish-chainsaw/handshake"
	"github.com/Richd0tcom/bookish-chainsaw/message"
	"github.com/Richd0tcom/bookish-chainsaw/peers"
)

type Client struct {
	Conn net.Conn

	// Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func shakeHands(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error){

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disables the deadline i.e it will not time out

	firstHand := handshake.New(infoHash, peerID)

	_, err := conn.Write(firstHand.Serialize())
	if err != nil {
		return nil, err
	}

	secondHand:= handshake.Handshake{}
	err = secondHand.Read(conn)

	if err != nil {
		return nil, err
	}

	if !bytes.Equal(secondHand.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x", secondHand.InfoHash, infoHash)
	}
	return &secondHand, nil

}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Expected bitfield but got %s", msg)
		return nil, err
	}
	if msg.ID != message.MSG_BITFIELD {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
func New(p peers.Peer, infoHash [20]byte, peerID [20]byte) (*Client, error) {
	// connect to peer
	conn, err := net.DialTimeout("tcp", p.String(), 3*time.Second)

	if err != nil {
		return nil, err
	}

    // send and receive handshake
	_, err= shakeHands(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// receive bitfield
	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     p,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil

}

// func (c *client) Piece(index int, begin, length int) []byte
// func (c *client) Cancel(index, begin, length int)



// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// Request sends a Request message to the peer
func (c *Client) Request(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// Interested sends an Interested message to the peer
func (c *Client) Interested() error {
	msg := message.Message{ID: message.MSG_INTERESTED}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// NotInterested sends a NotInterested message to the peer
func (c *Client) NotInterested() error {
	msg := message.Message{ID: message.MSG_NOT_INTERESTED}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// Unchoke sends an Unchoke message to the peer
func (c *Client) Unchoke() error {
	msg := message.Message{ID: message.MSG_UNCHOKE}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

