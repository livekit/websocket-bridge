package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/at-wat/ebml-go"
	"github.com/pion/webrtc/v3/pkg/media"
	"golang.org/x/net/websocket"
)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	host      = flag.String("host", "", "livekit server host")
	apiKey    = flag.String("api-key", "", "livekit api key")
	apiSecret = flag.String("api-secret", "", "livekit api secret")
	roomName  = flag.String("room-name", "", "room name")
	identity  = flag.String("identity", "", "participant identity")

	clusterId = []byte{0x1F, 0x43, 0xB6, 0x75}
)

func websocketServer(ws *websocket.Conn) {
	websocketBuff := make([]byte, 5*1000*1000)
	mkvBuff := []byte{}
	foundClusters := false

	videoTrack := connectionToLiveKit()

	var (
		block struct {
			Block ebml.Block `ebml:"SimpleBlock,stop"`
		}

		timestamp struct {
			Timestamp uint64 `ebml:"Timestamp,stop"`
		}

		bytesBuffer                         bytes.Buffer
		timestampOffset, lastVideoTimestamp int64
	)

	for {
		n, err := ws.Read(websocketBuff)
		if err != nil {
			panic(err)
		}
		mkvBuff = append(mkvBuff, websocketBuff[:n]...)

		if !foundClusters {
			clusterIndex := bytes.Index(mkvBuff, clusterId)
			if clusterIndex == -1 {
				continue
			}

			mkvBuff = mkvBuff[clusterIndex:]
			foundClusters = true
		}

		for {
			if clusterIndex := bytes.Index(mkvBuff, clusterId); clusterIndex == 0 {
				mkvBuff = mkvBuff[len(clusterId)+8:] // Remove Cluster ID
				if err = ebml.Unmarshal(bytes.NewReader(mkvBuff), &timestamp); err != nil && !errors.Is(err, ebml.ErrReadStopped) {
					panic(err)
				} else if err = ebml.Marshal(&timestamp, &bytesBuffer); err != nil {
					panic(err)
				}

				timestampOffset = int64(timestamp.Timestamp)
				mkvBuff = mkvBuff[bytesBuffer.Len():]
				bytesBuffer.Reset()
			}

			currentTimestamp := int64(block.Block.Timecode) + int64(timestampOffset)
			millisecondDiff := currentTimestamp - lastVideoTimestamp
			lastVideoTimestamp = currentTimestamp

			if err = ebml.Unmarshal(bytes.NewReader(mkvBuff), &block); errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} else if err != nil && !errors.Is(err, ebml.ErrReadStopped) {
				panic(err)
			} else if len(block.Block.Data) != 1 {
				panic("Unexpected Block Data Length")
			}

			if block.Block.TrackNumber == 2 {
				if err = videoTrack.WriteSample(media.Sample{Data: block.Block.Data[0], Duration: time.Duration(millisecondDiff) * time.Millisecond}, nil); err != nil {
					panic(err)
				}
			}

			lengthSize := variableWidthUintSize(uint64(len(block.Block.Data[0])))
			mkvBuff = mkvBuff[lengthSize+5:]             // Remove Block Header
			mkvBuff = mkvBuff[len(block.Block.Data[0]):] // Remove Block Data
		}
	}
}

func main() {
	flag.Parse()

	if *host == "" || *apiKey == "" || *apiSecret == "" || *roomName == "" || *identity == "" {
		panic("All CLI flags must be specified")
	}

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/websocket", websocket.Handler(websocketServer))

	fmt.Printf("Open %s to access this demo\n", *addr)
	panic(http.ListenAndServe(*addr, nil))
}
