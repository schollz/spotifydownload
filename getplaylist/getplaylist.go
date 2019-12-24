package getplaylist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/schollz/logger"
)

// Track is the basic track entity
type Track struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// getStringInBetween Returns empty string if no start string found
func getStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if s+e < len(str) && e > 0 {
		result = str[s : s+e]
	}
	return
}

type SpotifyData struct {
	Description string   `json:"description"`
	Href        string   `json:"href"`
	ID          string   `json:"id"`
	Images      []Images `json:"images"`
	Name        string   `json:"name"`
	Owner       Owner    `json:"owner"`
	Public      bool     `json:"public"`
	SnapshotID  string   `json:"snapshot_id"`
	Tracks      Tracks   `json:"tracks"`
	Type        string   `json:"type"`
	URI         string   `json:"uri"`
}
type Images struct {
	URL string `json:"url"`
}
type Owner struct {
	DisplayName string `json:"display_name"`
	Href        string `json:"href"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
}
type AddedBy struct {
	Href string `json:"href"`
	ID   string `json:"id"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}
type Artists struct {
	Href string `json:"href"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}
type Album struct {
	AlbumType            string    `json:"album_type"`
	Artists              []Artists `json:"artists"`
	Href                 string    `json:"href"`
	ID                   string    `json:"id"`
	Images               []Images  `json:"images"`
	Name                 string    `json:"name"`
	ReleaseDate          string    `json:"release_date"`
	ReleaseDatePrecision string    `json:"release_date_precision"`
	TotalTracks          int       `json:"total_tracks"`
	Type                 string    `json:"type"`
	URI                  string    `json:"uri"`
}
type ExternalIds struct {
	Isrc string `json:"isrc"`
}
type SpotifyTrack struct {
	Album       Album       `json:"album"`
	Artists     []Artists   `json:"artists"`
	DiscNumber  int         `json:"disc_number"`
	DurationMs  int         `json:"duration_ms"`
	Episode     bool        `json:"episode"`
	Explicit    bool        `json:"explicit"`
	ExternalIds ExternalIds `json:"external_ids"`
	Href        string      `json:"href"`
	ID          string      `json:"id"`
	IsLocal     bool        `json:"is_local"`
	IsPlayable  bool        `json:"is_playable"`
	Name        string      `json:"name"`
	Popularity  int         `json:"popularity"`
	PreviewURL  string      `json:"preview_url"`
	TrackNumber int         `json:"track_number"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
}
type LinkedFrom struct {
	Href string `json:"href"`
	ID   string `json:"id"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}
type Items struct {
	AddedAt time.Time    `json:"added_at"`
	AddedBy AddedBy      `json:"added_by"`
	IsLocal bool         `json:"is_local"`
	Track   SpotifyTrack `json:"track,omitempty"`
}
type Tracks struct {
	Href   string  `json:"href"`
	Items  []Items `json:"items"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
}

// GetTracks will return the playlist name and list of tracks from a Spotify playlist
func GetTracks(spotifyURL string) (playlistName string, tracks []Track, err error) {
	accessToken, err := getAccessToken(spotifyURL)
	if err != nil {
		return
	}

	foo := strings.Split(spotifyURL, "/playlist/")
	if len(foo) < 2 {
		err = fmt.Errorf("could not get id")
		return
	}
	playlistID := strings.Split(foo[1], "/")[0]
	playlistID = strings.Split(playlistID, "?")[0]
	log.Tracef("playlistID: '%s'", playlistID)
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/playlists/"+playlistID+"?type=track%2Cepisode&market=US", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:71.0) Gecko/20100101 Firefox/71.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Referer", "https://open.spotify.com/")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data SpotifyData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		err = errors.Wrap(err, "could not decode spotify data")
	}

	log.Tracef("data: %+v", data)
	tracks = make([]Track, len(data.Tracks.Items))
	if len(tracks) == 0 {
		err = fmt.Errorf("could not find any tracks")
		return
	}
	for i, track := range data.Tracks.Items {
		tracks[i] = Track{
			Number: i,
			Title:  track.Track.Name,
			Artist: track.Track.Artists[0].Name,
		}
	}
	playlistName = data.Name
	return
}

func getAccessToken(spotifyURL string) (accessToken string, err error) {
	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		log.Error(err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:71.0) Gecko/20100101 Firefox/71.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "sp_ab=%7B%7D; sp_landing=http%3A%2F%2Fopen.spotify.com%2Fplaylist%2F37i9dQZF1EtsXGZhBtSWWl; sp_t=c695ff90921aafb17baa61ea6c01c2f8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	accessToken = getStringInBetween(string(bodyBytes), `"accessToken":"`, `"`)
	log.Tracef("got access token: %s", accessToken)

	if len(accessToken) < 3 {
		err = fmt.Errorf("got no access token")
	}
	return
}
