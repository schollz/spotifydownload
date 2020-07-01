package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/schollz/getsong"
	log "github.com/schollz/logger"
	"github.com/schollz/spotifydownload/getplaylist"
)

type Result struct {
	err   error
	track getplaylist.Track
}

var debug, verbose bool

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func main() {
	var playlistURL string
	flag.StringVar(&playlistURL, "playlist", "", "The Spotify playlist URL link")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	if debug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("warn")
	}
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
		log.Infof("\nProblem with your bearer key or your playlist ID: %s\n", err.Error())
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

	runtime.GOMAXPROCS(4 * runtime.NumCPU())
	workers := runtime.NumCPU()

	tracksToDownload := make([]getplaylist.Track, len(tracks))
	i := 0
	for _, track := range tracks {
		fname := fmt.Sprintf("%s - %s.m4a", track.Artist, track.Title)
		if _, err = os.Stat(fname); os.IsNotExist(err) {
			tracksToDownload[i] = track
			i++
		}
	}
	tracksToDownload = tracksToDownload[:i]
	if len(tracksToDownload) == len(tracks) {
		log.Infof("Downloading %d tracks to '%s' folder\n", len(tracks), playlistName)
	} else {
		log.Infof("Downloading remaining %d of %d tracks to '%s' folder\n", len(tracksToDownload), len(tracks), playlistName)
	}

	jobs := make(chan getplaylist.Track, len(tracksToDownload))

	results := make(chan Result, len(tracksToDownload))

	for w := 0; w < workers; w++ {
		go func(jobs <-chan getplaylist.Track, results chan<- Result) {
			for j := range jobs {
				_, err := getsong.GetSong(j.Title, j.Artist, getsong.Options{
					ShowProgress: false, //debug,
					Debug:        debug,
					Filename:     fmt.Sprintf("%s - %s", j.Artist, j.Title),
					// DoNotDownload: true,
				})
				if err == nil {
					log.Infof("'%s' by '%s' downloaded.", j.Title, j.Artist)
				} else {
					log.Warnf("'%s' by '%s' not downloaded: %s.", j.Title, j.Artist, err.Error())
				}
				results <- Result{
					track: j,
					err:   err,
				}
			}
		}(jobs, results)
	}

	for _, track := range tracksToDownload {
		jobs <- track
	}
	close(jobs)

	for i := 0; i < len(tracksToDownload); i++ {
		<-results
	}
	err = nil
	return
}
