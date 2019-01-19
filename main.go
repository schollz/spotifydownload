package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/schollz/getsong"
)

type Result struct {
	err   error
	track Track
}

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
	trackNum := 1
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		if strings.Contains(line, `Spotify.Entity =`) {
			data := strings.TrimSpace(strings.Split(line, `Spotify.Entity =`)[1])
			data = data[:len(data)-1]
			var spotifyJSON SpotifyTracks
			err = json.Unmarshal([]byte(data), &spotifyJSON)
			playlistName = spotifyJSON.Name
			if err == nil {
				for _, track := range spotifyJSON.Tracks.Items {
					tracks = append(tracks, Track{trackNum, track.Track.Name, track.Track.Artists[0].Name})
					trackNum += 1
				}
			}
		}
	}
	return
}

var debug, verbose bool

func main() {
	var playlistURL string
	flag.StringVar(&playlistURL, "playlist", "", "The Spotify playlist URL link")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&verbose, "verbose", false, "Verbose mode")
	flag.Parse()

	if playlistURL == "" {
		fmt.Print(`

Enter your Spotify Playlist link.

To get the Spotify URL link you can right click on the playlist. 
If you are using the Desktop client, then you'll see a button 
"Shared > 🔗 Copy Playlist Link", or in the Web browser 
you'll see "Copy Playlist Link".

Enter the Spotify Playlist link here:
		
`)
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

	workers := 1

	tracksToDownload := make([]string, len(tracks))

	jobs := make(chan Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan Track, results chan<- Result) {
			for j := range jobs {
				fmt.Printf("%2d) Getting '%s' by '%s'..", j.Number, j.Title, j.Artist)
				_, err := getsong.GetSong(j.Title, j.Artist, getsong.Options{
					ShowProgress: debug,
					Debug:        debug,
					// DoNotDownload: true,
				})
				if err != nil {
					fmt.Printf("..error: %s\n", err)
				} else {
					fmt.Printf("..done.\n")
				}

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

	fmt.Printf("Downloading %d tracks in the '%s' playlist\n", len(tracks), playlistName)
	for i := 0; i < len(tracksToDownload); i++ {
		<-results
	}
	return
}
