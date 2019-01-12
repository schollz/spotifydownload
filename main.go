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

type Spotify struct {
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
		IsLocal      bool         `json:"is_local"`
		PrimaryColor *interface{} `json:"primary_color"`
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
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
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
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Episode          bool     `json:"episode"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string `json:"href"`
			ID          string `json:"id"`
			IsLocal     bool   `json:"is_local"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			Track       bool   `json:"track"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
		} `json:"track"`
		VideoThumbnail struct {
			URL *interface{} `json:"url"`
		} `json:"video_thumbnail"`
	} `json:"items"`
	Limit    int          `json:"limit"`
	Next     *interface{} `json:"next"`
	Offset   int          `json:"offset"`
	Previous *interface{} `json:"previous"`
	Total    int          `json:"total"`
}

type Result struct {
	err   error
	track Track
}

type Track struct {
	title    string
	artist   string
	duration int
}

func main() {
	var playlistID, bearerToken string
	flag.StringVar(&playlistID, "playlist", "", "The Spotify playlist ID")
	flag.StringVar(&bearerToken, "bearer", "", "Bearer token to use for Spotify")
	flag.Parse()

	if bearerToken == "" {
		fmt.Print(`Go to the Spotify Developer page: https://developer.spotify.com/console/get-playlist-tracks

At the bottom click "Get Token" and choose the playlist permissions 
"playlist-read-private" and "playlist-read-collaborative". 
Press "Request Token" and you'll be redirected to a Sign-in page.

Sign in with your credentials and then you'll be redirected back to 
the Spotify Developer page. Your Bearer key is at the bottom 
under "OAuth Token".

Copy that bearer token here: `)
		reader := bufio.NewReader(os.Stdin)
		bearerToken, _ = reader.ReadString('\n')
		bearerToken = strings.TrimSpace(bearerToken)
	}

	if playlistID == "" {
		fmt.Print(`To get the playlist tracks, you'll first need the playlist ID. 
To get the playlist ID, just right-click the playlist and goto 
Share -> Playlist URL. The URL will be something like

https://open.spotify.com/user/X/playlist/Y?si=Z
		
The playlist ID you need is "Y." Enter that playlist ID here: `)
		reader := bufio.NewReader(os.Stdin)
		playlistID, _ = reader.ReadString('\n')
		playlistID = strings.TrimSpace(playlistID)
	}

	err := run(bearerToken, playlistID)
	if err != nil {
		fmt.Printf("\nProblem with your bearer key or your playlist ID: %s\n", err.Error())
	}
}

func run(bearerToken, playlistID string) (err error) {
	getsong.OptionShowProgressBar = false

	spotifyJSON, err := getSpotifyPlaylist(bearerToken, playlistID)
	if err != nil {
		return
	}
	if len(spotifyJSON.Items) == 0 {
		err = fmt.Errorf("found no tracks")
		return
	}

	os.Mkdir(playlistID, 0644)
	os.Chdir(playlistID)

	workers := 30

	tracksToDownload := make([]string, len(spotifyJSON.Items))

	jobs := make(chan Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan Track, results chan<- Result) {
			for j := range jobs {
				var err error

				id, err := getsong.GetMusicVideoID(j.title+" "+j.artist, j.duration)
				if err == nil {
					fmt.Printf("Downloading %s by %s...\n", j.title, j.artist)
					fname, err := getsong.DownloadYouTube(id, j.artist+" - "+j.title)
					if err == nil {
						fmt.Printf("Converting %s by %s...\n", j.title, j.artist)
						err = getsong.ConvertToMp3(fname)
						if err == nil {
							fmt.Printf("Finished %s by %s...\n", j.title, j.artist)
						}
					}
				}

				results <- Result{
					track: j,
					err:   err,
				}
			}
		}(jobs, results)
	}

	for _, song := range spotifyJSON.Items {
		jobs <- Track{
			title:    song.Track.Name,
			artist:   song.Track.Artists[0].Name,
			duration: int(song.Track.DurationMs / 1000),
		}
	}
	close(jobs)

	for i := 0; i < len(tracksToDownload); i++ {
		result := <-results
		if result.err != nil {
			fmt.Printf("Error with %s by %s: %s\n", result.track.title, result.track.artist, result.err)
		} else {
			fmt.Printf("Finished %s by %s\n", result.track.title, result.track.artist)
		}
	}
	return
}

func getSpotifyPlaylist(bearerToken, playlistID string) (spotifyJSON Spotify, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID), nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &spotifyJSON)
	return
}
