<p align="center">
<img
    src=""
    width="260" border="0" alt="spotifydownload">
<br>
<a href="https://travis-ci.org/schollz/spotifydownload"><img
src="https://img.shields.io/travis/schollz/spotifydownload.svg?style=flat-square"
alt="Build Status"></a> <a
href="https://github.com/schollz/spotifydownload/releases/latest"><img
src="https://img.shields.io/badge/version-0.1.0-brightgreen.svg?style=flat-square"
alt="Version"></a> </p>

<p align="center">Automatically download your Spotify playlists.</p>

*spotifydownload* is an [open-source](https://github.com/schollz/spotifydownload) tool that makes it easy to download your Spotify playlists. It works by using a Bearer token to get the playlist from Spotify and then uses [getsong](https://github.com/schollz/getsong) to find the corect song and download it and convert it to an mp3.


# Install

Get the [latest release](https://github.com/schollz/spotifydownload/releases/latest) or install with `go get`:

```
go get github.com/schollz/spotifydownload
```

# Usage

## Basic usage

To run simply do

```bash
$ spotifydownload
```

and you'll be prompted to enter in your Bearer key and your Spotify playlist ID. If you already know your Bearer key and playlist ID you can enter these

```bash
$ spotifydownload -bearer BEARER -playlist PLAYLIST
```

## Advanced usage

If you want to automatically get the Bearer token you will need to install `puppeteer` in the folder you are going to run the program:

```
npm i puppeteer
```

And then add a `creds.js` file into the same folder:

```
module.exports = {
    username: 'SPOTIFY_USERNAME',
    password: 'SPOTIFY_PASSWORD'
}
```

and the program will automatically get the Bearer key for you when you run:

```bash
$ spotifydownload -playlist PLAYLIST
```


## Todo

- [ ] Remove leftover .webm things in the folder
- [ ] Store `index.js` in the code and write it to disk whenever it will be used
- [x] Store bearer key for multiple uses, and discard when it stops working
- [x] Check the api for the update getsong

## Contributing

Pull requests are welcome. Feel free to...

- Revise documentation
- Add new features
- Fix bugs
- Suggest improvements


## License

MIT

