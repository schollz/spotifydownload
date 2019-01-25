package main

import (
	"os"
	"path"
	"testing"

	"github.com/schollz/spotifydownload/getplaylist"
	"github.com/stretchr/testify/assert"
)

func TestGetTracks(t *testing.T) {
	spotifyURL := "https://open.spotify.com/user/spotify/playlist/37i9dQZEVXbrgWTCKQ0E8A?si=l5Pk_MH6TjOpKOUNhVm_zg"
	playlistName, tracks, err := getplaylist.GetTracks(spotifyURL)
	assert.Nil(t, err)
	assert.Equal(t, 30, len(tracks))
	assert.Equal(t, "Release Radar", playlistName)
}

func TestRun(t *testing.T) {
	spotifyURL := "https://open.spotify.com/user/123394108/playlist/6rEgbYZUO5yQ11zfg5NGac?si=zeHdEBJ_Rmui4ArUaSd-FQ"
	// os.RemoveAll("TestPlaylist")
	assert.Nil(t, run(spotifyURL))
	assert.True(t, exists("TestPlaylist"))
	assert.True(t, exists(path.Join("TestPlaylist", "Allen Toussaint - Old Records (oa6KzRfvtAs).mp3")))
	assert.True(t, exists(path.Join("TestPlaylist", "HAERTS - Eva (qxiOMm_x3Xg).mp3")))
}

// exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
