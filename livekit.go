package main

import (
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v3"
)

func connectionToLiveKit() *lksdk.LocalTrack {
	publicationOptions := &lksdk.TrackPublicationOptions{
		VideoWidth:  640,
		VideoHeight: 480,
		Name:        "webcam",
	}

	room, err := lksdk.ConnectToRoom(*host, lksdk.ConnectInfo{
		APIKey:              *apiKey,
		APISecret:           *apiSecret,
		RoomName:            *roomName,
		ParticipantIdentity: *identity,
	}, &lksdk.RoomCallback{}, lksdk.WithAutoSubscribe(false))
	if err != nil {
		panic(err)
	}

	videoTrack, err := lksdk.NewLocalTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8})
	if err != nil {
		panic(err)
	}

	if _, err = room.LocalParticipant.PublishTrack(videoTrack, publicationOptions); err != nil {
		panic(err)
	}

	return videoTrack
}
