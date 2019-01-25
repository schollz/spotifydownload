package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/getsong"
	"github.com/schollz/spotifydownload/getplaylist"
)

type Result struct {
	err   error
	track getplaylist.Track
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
	playlistName, tracks, err := getplaylist.GetTracks(playlistURL)
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

	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return
	}
	err = os.Chdir(playlistName)
	defer os.Chdir(cwd)
	if err != nil {
		return
	}

	bJSON, _ := json.MarshalIndent(tracks, "", " ")
	ioutil.WriteFile(time.Now().Format("2006-01-02")+".json", bJSON, 0644)

	workers := 1

	tracksToDownload := make([]string, len(tracks))

	jobs := make(chan getplaylist.Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan getplaylist.Track, results chan<- Result) {
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
