package getplaylist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTracks(t *testing.T) {
	spotifyURL := "https://open.spotify.com/user/spotify/playlist/37i9dQZEVXbrgWTCKQ0E8A?si=l5Pk_MH6TjOpKOUNhVm_zg"
	playlistName, tracks, err := GetTracks(spotifyURL)
	assert.Nil(t, err)
	assert.Equal(t, 30, len(tracks))
	assert.Equal(t, "Release Radar", playlistName)
}

func TestGetTrack(t *testing.T) {
	artist, title, err := GetTrack("https://open.spotify.com/track/4ehAqz41dlWdLbBBIubIla")
	assert.Nil(t, err)
	assert.Equal(t, "Love Is Here To Stay - Recorded At Spotify Studios NYC", title)
	assert.Equal(t, "Tony Bennett", artist)
}
