<p align="center">
<img
    src="https://i.imgur.com/IjoGFnO.png"
    width="260" border="0" alt="spotifydownload">
<br>
<a href="https://travis-ci.org/schollz/spotifydownload"><img
src="https://img.shields.io/travis/schollz/spotifydownload.svg?style=flat-square"
alt="Build Status"></a> <a
href="https://github.com/schollz/spotifydownload/releases/latest"><img
src="https://img.shields.io/badge/version-0.2.0-brightgreen.svg?style=flat-square"
alt="Version"></a> </p>

<p align="center">Automatically download your Spotify playlists.</p>

*spotifydownload* is an [open-source](https://github.com/schollz/spotifydownload) tool that makes it easy to download your Spotify playlists, using [getsong](https://github.com/schollz/getsong) to find the corect song and download it and convert it to an mp3.


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

and you'll be prompted to enter your Spotify playlist URL. If you already know your playlist URL you can enter it:

```bash
$ spotifydownload -playlist PLAYLIST_URL
```

Now you can easily schedule this to run using `crontab`, just edit it with `crontab -e` and add the line:

```
0 0 * * 0 cd /folder/to/spotifydownload &&  ./spotifydownload --playlist PLAYLIST_URL
```

which will execute it every 7 days so that you will never lose any songs in your Release Radar or Discover Weekly.

## Contributing

Pull requests are welcome. Feel free to...

- Revise documentation
- Add new features
- Fix bugs
- Suggest improvements


## License

MIT

