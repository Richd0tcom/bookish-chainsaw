package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Peer encodes connection information for a peer
type Peer struct {
    IP   net.IP
    Port uint16
}
//peers are 6bytes long
const peerSize int = 6 

func ParsePeers (peersListBin []byte) ([]Peer, error){

    if len(peersListBin)% peerSize != 0 {
        err := fmt.Errorf("received malformed peers")
        return nil, err
    }

    numPeers := len(peersListBin) / peerSize

    peers:= make([]Peer, numPeers)

    for i:=0; i< numPeers; i++ {
        offset := i * peerSize

        peers[i].IP = net.IP(peersListBin[offset: offset+ 4]) //IP address ocupies the first six bytes

        //the port occupies the next 2 bytes in BigEndian
        peers[i].Port = binary.BigEndian.Uint16(peersListBin[offset+4: offset+6]) 
    }
    
    return peers,  nil
}

func StartConnection(ps []Peer) {
    for _, p := range ps {
        go func (p *Peer) {
            net.DialTimeout("tcp", p.IP.String(), 3*time.Second)
        }(&p)
    }
}

func (p Peer) String() string {
    return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}