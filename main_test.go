package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTracks(t *testing.T) {
	spotifyURL := "https://open.spotify.com/user/spotify/playlist/37i9dQZEVXbrgWTCKQ0E8A?si=l5Pk_MH6TjOpKOUNhVm_zg"
	playlistName, tracks, err := getTracks(spotifyURL)
	assert.Nil(t, err)
	assert.Equal(t, 30, len(tracks))
	assert.Equal(t, "Release Radar", playlistName)
	fmt.Println(tracks[0])
}
