package getplaylist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	log "github.com/schollz/logger"
)

func TestGetTracks(t *testing.T) {
	spotifyURL := "https://open.spotify.com/user/spotify/playlist/37i9dQZEVXbrgWTCKQ0E8A?si=l5Pk_MH6TjOpKOUNhVm_zg"
	playlistName, tracks, err := GetTracks(spotifyURL)
	assert.Nil(t, err)
	log.Infof("tracks: %+v",tracks)
	assert.Equal(t, 30, len(tracks))
	assert.Equal(t, "Release Radar", playlistName)
}
