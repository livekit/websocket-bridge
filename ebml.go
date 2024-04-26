package main

import (
	"bytes"
	"errors"

	"github.com/at-wat/ebml-go"
)

var clusterId = []byte{0x1F, 0x43, 0xB6, 0x75}

func variableWidthUintSize(v uint64) int {
	switch {
	case v < 0x80-1:
		return 1
	case v < 0x4000-1:
		return 2
	case v < 0x200000-1:
		return 3
	case v < 0x10000000-1:
		return 4
	case v < 0x800000000-1:
		return 5
	case v < 0x40000000000-1:
		return 6
	case v < 0x2000000000000-1:
		return 7
	default:
		return 8
	}
}

func readClusterHeader(mkvBuff []byte) (timestampOffset int64, amountRead int, err error) {
	if clusterIndex := bytes.Index(mkvBuff, clusterId); clusterIndex == 0 {
		var (
			bytesBuffer bytes.Buffer
			timestamp   struct {
				Timestamp uint64 `ebml:"Timestamp,stop"`
			}
		)

		amountRead = len(clusterId) + 8 // Remove Cluster ID
		if err = ebml.Unmarshal(bytes.NewReader(mkvBuff[amountRead:]), &timestamp); err != nil && !errors.Is(err, ebml.ErrReadStopped) {
			return
		} else if err = ebml.Marshal(&timestamp, &bytesBuffer); err != nil {
			return
		}

		timestampOffset = int64(timestamp.Timestamp)
		amountRead += bytesBuffer.Len()
	}

	return
}
