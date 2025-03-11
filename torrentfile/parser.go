package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	//pieces of the file
	Pieces      string `bencode:"pieces"` 

    PieceLength int    `bencode:"piece length"`
    Length      int    `bencode:"length"`
    Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string `bencode:"announce"`
	Info bencodeInfo `bencode:"info"`

}

//parses a bencoded torrent file
func OpenFile(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.parseToTorrentFile()
}

type TorrentFile struct {
    Announce    string
	//SHA-1 hash of the file (helps us know we're getting the right file)
    InfoHash    [20]byte

	//slice of hashes for each piece of the file
    PieceHashes [][20]byte
    PieceLength int

	//total length of the bytes of the file 
    Length      int
    Name        string
}

const Port uint16 = 1738

// returns the SHA-1 hash of the info that identifies the file
func (bi *bencodeInfo) hashInfo() ([20]byte, error){
	buffer:= bytes.Buffer{}

	err:= bencode.Marshal(&buffer, *bi)

	if err != nil {
		return [20]byte{}, err
	}
	
	hash:= sha1.Sum(buffer.Bytes())

	return hash, nil
} 

func (bi *bencodeInfo) splitPieceHash() ([][20]byte, error) {
	pieces:= []byte(bi.Pieces)
	singlePieceLength:= 20
	if len(pieces) % singlePieceLength != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(pieces))
		return nil, err
	}

	numberOfPieces := len(pieces) / singlePieceLength

	pieceHashes:= make([][20]byte, numberOfPieces)

	for  i:= 0; i<numberOfPieces; i++ {
		copy(pieceHashes[i][:], pieces[i * singlePieceLength : (i+1)* singlePieceLength])
	}

	return pieceHashes, nil

}

func (bto bencodeTorrent) parseToTorrentFile() (TorrentFile, error) {

	pieceHashes, err:= bto.Info.splitPieceHash()

	if err != nil {
		return TorrentFile{}, err
	}

	infoHash, err := bto.Info.hashInfo()

	if err != nil {
		return TorrentFile{}, err
	}
	
	torrentfile:= TorrentFile{
		Announce: bto.Announce,
		InfoHash: infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length: bto.Info.Length,
		Name: bto.Info.Name,
	}

	return torrentfile, nil
}





