package handshake

import (
	"fmt"
	"io"
)

//
//
//
//
//The protocol identifier, called the pstr which is always BitTorrent protocol
//Eight reserved bytes, all set to 0. We’d flip some of them to 1 to indicate that we support certain extensions. But we don’t, so we’ll keep them at 0.
//The infohash that we calculated earlier to identify which file we want
// The Peer ID that we made up to identify ourselves
//
//
//

//A bittorrent handshake is a special message that a peer uses to identify itself
type Handshake struct {
	Pstr string //The protocol identifier, called the pstr which is always: "Bittorrent Protocol"
	InfoHash [20]byte
	PeerID [20]byte
}

const BYTE_LEN = 49

// creates a new hand shake
func New(infoHash, peerID [20]byte) (*Handshake) {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (hs *Handshake) Serialize() []byte {

	// 1 byte for protocol id length
	// n bytes for protocol string 
	// 8 reserved bytes
	// 20 bytes for peer id
	// 20 bytes for info hash
	// 1 + 8 + 20 + 20 + n = 49 + n
	

	buf := make([]byte, len(hs.Pstr)+BYTE_LEN) 
	buf[0] = byte(len(hs.Pstr))
	offset :=1
	offset+= copy(buf[offset:], hs.Pstr)
	offset+= copy(buf[offset:], make([]byte, 8))
	offset+= copy(buf[offset:], hs.InfoHash[:])
	offset+= copy(buf[offset:], hs.PeerID[:])

	return buf
}

func (hs *Handshake) DeSerialize(pstrLen int, buf []byte) error {
	
	hs.Pstr = string(buf[0:pstrLen])
	hs.InfoHash = [20]byte(buf[8 + pstrLen: 20 +1])
	hs.PeerID = [20]byte(buf[8 + pstrLen + 20: ])

	return nil
}

func (hs *Handshake) Read(r io.Reader) error {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return err
	}

	handshakeBuf := make([]byte, BYTE_LEN-1+pstrlen) //we subtract one byte because we already read it
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return err
	}
	err = hs.DeSerialize(pstrlen, handshakeBuf)
	if err != nil {
		return err
	}
	return nil
}