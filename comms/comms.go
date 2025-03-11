package comms

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/Richd0tcom/bookish-chainsaw/comms/client"
	"github.com/Richd0tcom/bookish-chainsaw/message"
	"github.com/Richd0tcom/bookish-chainsaw/peers"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384 //16KB

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5


type Torrent struct {
	PieceHashes [][20]byte 
	InfoHash   [20]byte
	PeerID [20]byte

	Peers []peers.Peer
	Length int //length of a piece
}

type pieceWork struct {
	Index int
	PieceHash [20]byte
	Length int
}

type pieceResult struct {
    index int
    buf   []byte
}

type pieceProgress struct {
    index      int
    client     *client.Client
    buf        []byte
    downloaded int
    requested  int
    backlog    int
}

func (state *pieceProgress) readMessage() error {
    msgg, err := state.client.Read() // this call blocks
    if err != nil {
		return err
	}

    if msgg == nil { // keep-alive
		return nil
	}
    switch msgg.ID {
		case message.MSG_UNCHOKE:
			state.client.Choked = false
		case message.MSG_CHOKE:
			state.client.Choked = true
		case message.MSG_HAVE:
			index, err := message.ParseHave(msgg)

			if err != nil {
				return err
			}
			state.client.Bitfield.SetPiece(index)
		case message.MSG_PIECE:
			n, err := message.ParsePiece(state.index, state.buf, msgg)
			if err != nil {
				return err
			}

			state.downloaded += n
			state.backlog--
    }
    return nil
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.Length
	end = min(begin + t.Length, t.Length)
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}



func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
    state := pieceProgress{
        index:  pw.Index,
        client: c,
        buf:    make([]byte, pw.Length),
    }

    // Setting a deadline helps get unresponsive peers unstuck.
    // 30 seconds is more than enough time to download a 262 KB piece
    c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
    defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

    for state.downloaded < pw.Length {
        // If unchoked, send requests until we have enough unfulfilled requests
        if !state.client.Choked {
            for state.backlog < MaxBacklog && state.requested < pw.Length {
                blockSize := min(pw.Length - state.requested, MaxBlockSize)

                err := c.Request(pw.Index, state.requested, blockSize)
                if err != nil {
                    return nil, err
                }
                state.backlog++
                state.requested += blockSize
            }
        }

        err := state.readMessage()
        if err != nil {
            return nil, err
        }
    }

    return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	//compare the hashes of downloaded piece and the Piece hash info
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.PieceHash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.Index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.New(peer, t.InfoHash, t.PeerID )
    if err != nil {
        log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
        return
    }

	defer c.Conn.Close()
    log.Printf("Completed handshake with %s\n", peer.IP)

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.Index) {
			workQueue<- pw // Put piece back on the queue
		}

		// Download the piece
        buf, err := attemptDownloadPiece(c, pw)
        if err != nil {
            log.Println("Exiting", err)
            workQueue <- pw // Put piece back on the queue
            return
        }

		err = checkIntegrity(pw, buf)
		if err != nil {
            log.Printf("Piece #%d failed integrity check\n", pw.Index)
            workQueue <- pw // Put piece back on the queue
            continue
        }

		c.SendHave(pw.Index)
        results <- &pieceResult{pw.Index, buf}
	}
} 


func (t *Torrent) Download() {


	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}


	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++
	}
	close(workQueue)
}
