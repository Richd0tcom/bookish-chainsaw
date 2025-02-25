package handshake


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

func (hs *Handshake) Serialize() []byte{

	// 1 byte for protocol id length
	// 8 reserved bytes
	// 20 bytes for peer id
	// 20 bytes for info hash
	// 1 + 8 + 20 + 20 = 49
	BYTE_LEN := 49

	buf := make([]byte, len(hs.Pstr)+BYTE_LEN) 
	buf[0] = byte(len(hs.Pstr))
	offset :=1
	offset+= copy(buf[offset:], hs.Pstr)
	offset+= copy(buf[offset:], make([]byte, 8))
	offset+= copy(buf[offset:], hs.InfoHash[:])
	offset+= copy(buf[offset:], hs.PeerID[:])

	return buf
}

func (hs *Handshake) DeSerialize(buf []byte) error {
	pstrLen := int(buf[0])
	hs.Pstr = string(buf[1:pstrLen+1])
	hs.InfoHash = [20]byte(buf[9 + pstrLen: 20 +1])
	hs.PeerID = [20]byte(buf[9 + pstrLen + 20: ])

	return nil
}