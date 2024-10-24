package torrentfile

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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


func (tf *TorrentFile) ConnectToPeers() (trackerResp, error) {
	//create a peerID (random)

	var peer_id [20]byte 
	_, err:= rand.Read(peer_id[:])


	
	url, err:= tf.buildTrackerURL(Port, peer_id)


	c:= http.Client{}

	response, err:= c.Get(url)
	if err != nil {
		fmt.Println(err)
		return trackerResp{}, err
	}

	defer response.Body.Close()

	trackRes:= trackerResp{}

	err = bencode.Unmarshal(response.Body, &trackRes)

	if err != nil {
		fmt.Println(err)
		return trackerResp{}, err
	}

	return trackRes, nil

	//we will update the function to return Peers instead

	
}