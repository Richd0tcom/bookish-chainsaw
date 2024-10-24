package peers

import "net"

// Peer encodes connection information for a peer
type Peer struct {
    IP   net.IP
    Port uint16
}

func parsePeers () {
	
}