package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"flag"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"bufio"
	"time"

	"github.com/schollz/getsong"
)

type Result struct {
	err   error
	track Track
}

type Track struct {
	Title    string
	Artist   string
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

type SpotifyTracks struct {
	Collaborative bool   `json:"collaborative"`
	Description   string `json:"description"`
	ExternalUrls  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  interface{} `json:"href"`
		Total int         `json:"total"`
	} `json:"followers"`
	Href   string `json:"href"`
	ID     string `json:"id"`
	Images []struct {
		Height int    `json:"height"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
	} `json:"images"`
	Name  string `json:"name"`
	Owner struct {
		DisplayName  string `json:"display_name"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"owner"`
	PrimaryColor interface{} `json:"primary_color"`
	Public       bool        `json:"public"`
	SnapshotID   string      `json:"snapshot_id"`
	Tracks       struct {
		Href  string `json:"href"`
		Items []struct {
			AddedAt time.Time `json:"added_at"`
			AddedBy struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"added_by"`
			IsLocal      bool        `json:"is_local"`
			PrimaryColor interface{} `json:"primary_color"`
			Track        struct {
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
				Episode     bool `json:"episode"`
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
				Track       bool   `json:"track"`
				TrackNumber int    `json:"track_number"`
				Type        string `json:"type"`
				URI         string `json:"uri"`
			} `json:"track"`
			VideoThumbnail struct {
				URL interface{} `json:"url"`
			} `json:"video_thumbnail"`
		} `json:"items"`
		Limit    int         `json:"limit"`
		Next     interface{} `json:"next"`
		Offset   int         `json:"offset"`
		Previous interface{} `json:"previous"`
		Total    int         `json:"total"`
	} `json:"tracks"`
	Type                 string `json:"type"`
	URI                  string `json:"uri"`
	Etag                 string `json:"etag"`
	LastCheckedTimestamp int    `json:"last-checked-timestamp"`
}

func getTracks(spotifyURL string) (playlistName string, tracks []Track, err error) {
	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	tracks = []Track{}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(bodyBytes))
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		if strings.Contains(line,`Spotify.Entity =`) {
			data := strings.TrimSpace(strings.Split(line,`Spotify.Entity =`)[1])
			data = data[:len(data)-1]
			var spotifyJSON SpotifyTracks
			err = json.Unmarshal([]byte(data),&spotifyJSON)
			playlistName = spotifyJSON.Name
			if err == nil {
				for _, track := range spotifyJSON.Tracks.Items {
					tracks = append(tracks,Track{track.Track.Name,track.Track.Artists[0].Name})
				}
			}
		}
	}
	return
}


func getBearerKeyUsingChromeHeadless() (bearer string, err error) {
	cmd := exec.Command("node", "index.js")
	var bearerBytes []byte
	bearerBytes, _ = cmd.CombinedOutput()
	bearer = strings.TrimSpace(string(bearerBytes))
	if strings.Contains(bearer, `Cannot find module 'puppeteer'`) {
		err = fmt.Errorf("need to install puppeteer")
	}
	return
}

var debug bool

func main() {
	var playlistURL string
	flag.StringVar(&playlistURL, "playlist", "", "The Spotify playlist URL link")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	if playlistURL == "" {
		fmt.Print(`To get the playlist tracks, you'll first need the playlist URL. 
To get the playlist URL, just right-click the playlist and goto 
Share -> Playlist URL. The URL will be something like

https://open.spotify.com/user/X/playlist/Y?si=Z
		
Enter that playlist URL here: `)
		reader := bufio.NewReader(os.Stdin)
		playlistURL, _ = reader.ReadString('\n')
		playlistURL = strings.TrimSpace(playlistURL)
	}

	err := run(playlistURL)
	if err != nil {
		fmt.Printf("\nProblem with your bearer key or your playlist ID: %s\n", err.Error())
	}
}

func run(playlistURL string) (err error) {
	playlistName, tracks, err := getTracks(playlistURL)
	if err != nil {
		return
	}

	if len(tracks) == 0 {
		err = fmt.Errorf("found no tracks")
		return
	}

	if _, err = os.Stat(playlistName); os.IsNotExist(err) {
		err = os.Mkdir(playlistName, 0755)
		if err != nil {
			return
		}
	}

	err = os.Chdir(playlistName)
	if err != nil {
		return
	}

	bJSON, _ := json.MarshalIndent(tracks, "", " ")
	ioutil.WriteFile(time.Now().Format("2006-01-02")+".json", bJSON, 0644)

	workers := 30

	tracksToDownload := make([]string, len(tracks))

	jobs := make(chan Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan Track, results chan<- Result) {
			for j := range jobs {
				fmt.Printf("Downloading %s by %s...\n", j.Title, j.Artist)
				fname, err := getsong.GetSong(j.Title, j.Artist, getsong.Options{
					ShowProgress: true,
					Debug:        debug,
					// DoNotDownload: true,
				})
				fmt.Printf("Downloaded %s.\n", fname)

				results <- Result{
					track: j,
					err:   err,
				}
			}
		}(jobs, results)
	}

	for _, track := range tracks {
		jobs <- track
	}
	close(jobs)

	for i := 0; i < len(tracksToDownload); i++ {
		result := <-results
		if result.err != nil {
			fmt.Printf("Error with %s by %s: %s\n", result.track.Title, result.track.Artist, result.err)
		} else {
			fmt.Printf("Finished %s by %s\n", result.track.Title, result.track.Artist)
		}
	}
	return
}
