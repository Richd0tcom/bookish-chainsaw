package torrentfile

import (
	"crypto/rand"
	_ "crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Richd0tcom/bookish-chainsaw/comms"
	"github.com/Richd0tcom/bookish-chainsaw/peers"
	"github.com/jackpal/bencode-go"
)

type trackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

//builds the tracker URL so we can connect the tracker and search for peers
func (tf *TorrentFile) buildTrackerURL(port uint16, peerID [20]byte) (string, error) {
	baseURL, err :=url.Parse(tf.Announce)

	if err != nil {
		return "", err
	}

	params := url.Values{
        "info_hash":  []string{string(tf.InfoHash[:])},
        "peer_id":    []string{string(peerID[:])},
        "port":       []string{strconv.Itoa(int(port))},
        "uploaded":   []string{"0"},
        "downloaded": []string{"0"},
        "compact":    []string{"1"},
        "left":       []string{strconv.Itoa(tf.Length)},
    }

	baseURL.RawQuery = params.Encode()
	return baseURL.String(), nil
}


func (tf *TorrentFile) ConnectToPeers(peerID [20]byte) ([]peers.Peer, error) {

	url, err:= tf.buildTrackerURL(Port, peerID)


	c:= http.Client{}

	response, err:= c.Get(url)
	if err != nil {
		fmt.Println(err)
		return []peers.Peer{}, err
	}

	defer response.Body.Close()

	trackRes:= trackerResp{}

	err = bencode.Unmarshal(response.Body, &trackRes)

	if err != nil {
		fmt.Println(err)
		return []peers.Peer{}, err
	}

	fmt.Println(trackRes)

	return peers.ParsePeers([]byte(trackRes.Peers))


	//we will update the function to return Peers instead

	
}

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	peers, err := t.ConnectToPeers(peerID)
	if err != nil {
		return err
	}

	torrent := comms.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}