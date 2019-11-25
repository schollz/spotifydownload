package getplaylist

import (
	"io/ioutil"
	"net/http"
	"strings"

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

type SpotifyTrack struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		TotalTracks          int    `json:"total_tracks"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	DiscNumber  int  `json:"disc_number"`
	DurationMs  int  `json:"duration_ms"`
	Explicit    bool `json:"explicit"`
	ExternalIds struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string `json:"href"`
	ID          string `json:"id"`
	IsLocal     bool   `json:"is_local"`
	IsPlayable  bool   `json:"is_playable"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
}

func GetArtist(spotifyURL string) (artistName string, err error) {
	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		if strings.Contains(line, `meta property="og:title" content="`) {
			data := strings.TrimSpace(strings.Split(line, `meta property="og:title" content="`)[1])
			artistName = strings.Split(data, `"`)[0]
			break
		}
	}
	return
}

func GetTrack(spotifyTrackURL string) (artistName, trackName string, err error) {
	log.Debugf("getting track: %s", spotifyTrackURL)
	req, err := http.NewRequest("GET", spotifyTrackURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		if strings.Contains(line, `meta property="og:title" content="`) {
			data := strings.TrimSpace(strings.Split(line, `meta property="og:title" content="`)[1])
			trackName = strings.Split(data, `"`)[0]
		}
		if strings.Contains(line, `meta property="music:musician" content="`) {
			data := strings.TrimSpace(strings.Split(line, `meta property="music:musician" content="`)[1])
			data = strings.Split(data, `"`)[0]
			artistName, err = GetArtist(data)
			break
		}
	}
	return
}

// GetTracks will return the playlist name and list of tracks from a Spotify playlist
func GetTracks(spotifyURL string) (playlistName string, tracks []Track, err error) {
	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	tracks = []Track{}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	trackNum := 1
	for _, line := range strings.Split(string(bodyBytes), "<") {
		if strings.Contains(line, `meta property="og:title" content="`) {
			data := strings.TrimSpace(strings.Split(line, `meta property="og:title" content="`)[1])
			data = strings.Split(data, `"`)[0]
			playlistName = strings.Split(data, ",")[0]
		}

		if strings.Contains(line, `meta property="music:song" content="`) {
			data := strings.TrimSpace(strings.Split(line, `meta property="music:song" content="`)[1])
			data = strings.Split(data, `"`)[0]
			log.Debugf("found track in playlist: %s", data)
			track := Track{Number: trackNum}
			track.Artist, track.Title, err = GetTrack(data)
			if err != nil {
				log.Debugf("error: %s", err)
				continue
			}
			tracks = append(tracks, track)
			trackNum++
		}
	}
	// fmt.Println(tracks)
	// if len(tracks) == 0 {
	// 	fmt.Println(spotifyURL)
	// 	fmt.Println(string(bodyBytes))
	// }
	return
}
