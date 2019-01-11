# -*- coding: utf-8 -*-

"""Main module."""
from __future__ import unicode_literals
from urllib.parse import quote
import json
import sys

import youtube_dl
import requests
from requests_html import HTMLSession


def find_between(s, first, last):
    try:
        start = s.index(first) + len(first)
        end = s.index(last, start)
        return s[start:end]
    except ValueError:
        return ""


class MyLogger(object):
    def debug(self, msg):
        print(msg)
        pass

    def warning(self, msg):
        pass

    def error(self, msg):
        print(msg)


def my_hook(d):
    print(d["status"])
    if d["status"] == "finished":
        print("Done downloading, now converting ...")


def download_youtube(youtubeID):
    ydl_opts = {
        "format": "bestaudio/best",
        "postprocessors": [
            {
                "key": "FFmpegExtractAudio",
                "preferredcodec": "mp3",
                "preferredquality": "192",
            }
        ],
        "logger": MyLogger(),
        "progress_hooks": [my_hook],
    }
    with youtube_dl.YoutubeDL(ydl_opts) as ydl:
        ydl.download(["https://www.youtube.com/watch?v={}".format(youtubeID)])


def find_youtube_id(searchTerm):
    session = HTMLSession()
    r = session.get(
        "https://www.youtube.com/results?search_query={}".format(quote(searchTerm))
    )
    jsonData = json.loads(
        find_between(
            r.text, 'ytInitialData"] =', 'window["ytInitialPlayerResponse'
        ).strip()[:-1]
    )

    ids = []
    videosData = jsonData["contents"]["twoColumnSearchResultsRenderer"][
        "primaryContents"
    ]["sectionListRenderer"]["contents"][0]["itemSectionRenderer"]["contents"]
    for videoData in videosData:
        description = ""
        if "videoRenderer" not in videoData:
            continue
        if "videoId" not in videoData["videoRenderer"]:
            continue
        if "descriptionSnippet" not in videoData["videoRenderer"]:
            continue
        if "runs" not in videoData["videoRenderer"]["descriptionSnippet"]:
            continue
        for run in videoData["videoRenderer"]["descriptionSnippet"]["runs"]:
            description += run["text"]
        print(description)
        if "Provided to YouTube" in description:
            ids.append(videoData["videoRenderer"]["videoId"])

    return ids


def download_song(song):
    try:
        ids = find_youtube_id(song)
        if len(ids) > 0:
            download_youtube(ids[0])
    except Exception as e:
        print(e)


def download_playlist(playlist):
    playlist_json = json.load(open(playlist, "rb"))
    for song in playlist_json["items"]:
        try:
            title = song["track"]["name"] + " " + song["track"]["artists"][0]["name"]
            download_song(title)
        except Exception as e:
            print(e)
