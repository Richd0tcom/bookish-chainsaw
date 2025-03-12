# bookish-chainsaw

Tiny BitTorrent client written in Go.

## Install

```sh
go get github.com/richd0tcom/bookish-chainsaw
```

## Usage
Try downloading [Debian](https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/#indexlist)!

```sh
bookish-chainsaw debian-10.2.0-amd64-netinst.iso.torrent debian.iso
```



## Limitations/TODO
* Only supports `.torrent` files (no magnet links)
* Based on the earliest specification of bittorrent (may not work with some modern torrent files)
* Only supports HTTP trackers
* Does not support multi-file torrents
* Strictly leeches (does not support uploading pieces)
