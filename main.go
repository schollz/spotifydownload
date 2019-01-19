package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
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

func getTracks(spotifyURL string) (tracks []Track, err error) {
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
	for _, line := range strings.Split(string(bodyBytes), "\n") {
		if strings.Contains(line, `tracklist-col name`) {
			artist := getStringInBetween(line, `<span dir="auto">`, `<`)
			foo := strings.Split(line, artist)
			if len(foo) < 2 {
				continue
			}
			title := getStringInBetween(foo[1], `<span dir="auto">`, `<`)
			tracks = append(tracks,Track{title,artist})
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
	var playlistID, bearerToken string
	flag.StringVar(&playlistID, "playlist", "", "The Spotify playlist ID")
	flag.StringVar(&bearerToken, "bearer", "", "Bearer token to use for Spotify")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	// first try to find saved bearer token
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	cacheFolder := path.Join(homeDir, ".cache", "spotifydownload")
	os.MkdirAll(cacheFolder, 0755)
	if _, err := os.Stat(path.Join(cacheFolder, "bearer.token")); err == nil {
		bearerBytes, errRead := ioutil.ReadFile(path.Join(cacheFolder, "bearer.token"))
		if errRead == nil {
			bearerToken = string(bearerBytes)
			_, errTestBearer := getCurrentPlaylists(bearerToken)
			if errTestBearer != nil {
				// bearer doesn't work, get another
				bearerToken = ""
			} else {
				fmt.Println("Loaded previous Bearer key")
			}
		}
	}

	_, credsErr := os.Stat("creds.js")
	if bearerToken == "" && credsErr == nil {
		var errBearerFromNode error
		bearerToken, errBearerFromNode = getBearerKeyUsingChromeHeadless()
		if errBearerFromNode != nil {
			bearerToken = ""
		} else {
			fmt.Println("Got Bearer key using node")
		}

	}

	_, errTestBearer := getCurrentPlaylists(bearerToken)
	if errTestBearer != nil {
		// bearer doesn't work, get another
		fmt.Println("Current bearer key has expired")
		bearerToken = ""
	}

	if bearerToken == "" {
		fmt.Print(`Go to the Spotify Developer page: 
		
		https://developer.spotify.com/console/get-playlist-tracks

		At the bottom click "Get Token" and choose the playlist permissions 
		"playlist-read-private" and "playlist-read-collaborative". 
		Press "Request Token" and you'll be redirected to a Sign-in page.
		
		Sign in with your credentials and then you'll be redirected back to 
		the Spotify Developer page. Your Bearer key is at the bottom 
		under "OAuth Token".

`)
		for {
			fmt.Println(`Copy that bearer token here: `)
			reader := bufio.NewReader(os.Stdin)
			bearerToken, _ = reader.ReadString('\n')
			bearerToken = strings.TrimSpace(bearerToken)
			_, errTestBearer := getCurrentPlaylists(bearerToken)
			if errTestBearer != nil {
				// bearer doesn't work, get another
				fmt.Println("Incorrect Bearer key")
				bearerToken = ""
			}
			if bearerToken != "" {
				break
			}
		}

	}

	// save correct Bearer key
	err = ioutil.WriteFile(path.Join(cacheFolder, "bearer.token"), []byte(bearerToken), 0755)
	if err != nil {
		panic(err)
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

	err = run(bearerToken, playlistID)
	if err != nil {
		fmt.Printf("\nProblem with your bearer key or your playlist ID: %s\n", err.Error())
	}
}

func run(playlistURL string) (err error) {
	playlistName, tracks, err := getTracks(playlistURL)
	if err != nil {
		return
	}

	if len(spotifyJSON.Tracks.Items) == 0 {
		err = fmt.Errorf("found no tracks")
		return
	}

	if _, err = os.Stat(spotifyJSON.Name); os.IsNotExist(err) {
		err = os.Mkdir(spotifyJSON.Name, 0755)
		if err != nil {
			return
		}
	}

	err = os.Chdir(spotifyJSON.Name)
	if err != nil {
		return
	}

	bJSON, _ := json.MarshalIndent(spotifyJSON, "", " ")
	ioutil.WriteFile(time.Now().Format("2006-01-02")+".json", bJSON, 0644)

	workers := 30

	tracksToDownload := make([]string, len(spotifyJSON.Tracks.Items))

	jobs := make(chan Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan Track, results chan<- Result) {
			for j := range jobs {
				fmt.Printf("Downloading %s by %s...\n", j.Title, j.Artist)
				fname, err := getsong.GetSong(j.Title, j.Artist, getsong.Options{
					// Duration:      j.duration,
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

	for _, song := range spotifyJSON.Tracks.Items {
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
			fmt.Printf("Error with %s by %s: %s\n", result.track.Title, result.track.Artist, result.err)
		} else {
			fmt.Printf("Finished %s by %s\n", result.track.Title, result.track.Artist)
		}
	}
	return
}

type Spotify struct {
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
		Height interface{} `json:"height"`
		URL    string      `json:"url"`
		Width  interface{} `json:"width"`
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
				URL interface{} `json:"url"`
			} `json:"video_thumbnail"`
		} `json:"items"`
		Limit    int         `json:"limit"`
		Next     interface{} `json:"next"`
		Offset   int         `json:"offset"`
		Previous interface{} `json:"previous"`
		Total    int         `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

func getSpotifyPlaylist(bearerToken, playlistID string) (spotifyJSON Spotify, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", playlistID), nil)
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

type UserPlaylists struct {
	Href  string `json:"href"`
	Items []struct {
		Collaborative bool `json:"collaborative"`
		ExternalUrls  struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height interface{} `json:"height"`
			URL    string      `json:"url"`
			Width  interface{} `json:"width"`
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
			Total int    `json:"total"`
		} `json:"tracks"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"items"`
	Limit    int         `json:"limit"`
	Next     string      `json:"next"`
	Offset   int         `json:"offset"`
	Previous interface{} `json:"previous"`
	Total    int         `json:"total"`
}

func getCurrentPlaylists(bearerToken string) (spotifyJSON UserPlaylists, err error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/playlists", nil)
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

	if resp.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &spotifyJSON)
	} else {
		err = fmt.Errorf("get basic playlists received %d", resp.StatusCode)
	}
	return
}
