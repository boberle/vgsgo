# VGSGO - Video Game Soundtrack Player

- Play music files in loop with loop start/end points,
- rate them,
- from a local or remote library.

## How to Use?

**Step 1:** Find some video game music (or other).

**Step 2:** Write a metadata JSON file, with the metadata of each music file, like this:

```json
[
  {
    "path": "path/to/file.brstm",
    "timestamp": 1699652599,
    "loop_start": 2687000,
    "loop_end": 0,
    "duration": 125.1,
    "size": 7425436,
    "title": "Song Title",
    "game_title": "Game Title",
    "error": false
  },
  ...
]
```

This file must be placed alongside the music files. For example, if `path` is `game_123/song_456.brstm`, then the tree should look like:

```
root/
├── game_123
│   └── song_456.brstm
└── metadata.json
```

**Step 3:** Build the program. Assuming you have go installed:

```bash
make vgsgo
# or:
go build -o build/vgsgo cmd/app/main.go
```

**Step 4:** Run the program.

```bash
./vgsgo /path/to/metadata.json
```

You can specify several metadata files.

Here are the switches and options:

- `-continuous`: don't stop to ask rating
- `-game-title STRING`: limit to song with a game title that contains the string
- `-max-plays INT`: maximum number of plays (default is 0, infinity)
- `-min-duration INT`: minimum duration
- `-min-rating FLOAT`: minimum rating. Add `--only-has-rating` to limit to songs that have ratings
- `-only-has-no-rating`: limit to songs that don't have a rating
- `-only-has-rating`: limit to songs that have a rating
- `-play-last`: don't shuffle songs, play the last ones
- `-rating-file STRING`: json file where ratings are store
- `-title string`: limit to song with a title that contains the string

**Step 5:** Play and rate the songs.

A song is chosen randomly (according to the filters you specified on the command line, e.g. `-min-rating`). `mplayer` will play the song, looping according to the loop start/end points defined in the metadata file. If `--max-plays` is set, then it will loop that maximum times, otherwise it will loop indefinitely. You can stop by pressing `q`.

Then you will be asked to enter the next action:

```
[<INT>][r][q]
```

- `INT` is the rating (between 1 and 5 incl.)
- `r` is to resume playing the song (indefinitely)
- `q` is to quit the program

If you enter nothing (or just a rating), the next song will (chosen randomly) will be played.

![demo](doc/demo.gif)

If you set a rating file with the `-rating-file` option, then the program creates a json file that looks like:

```json
[
  {
    "path": "path/to/song.brstm",
    "plays": [
      {
        "timestamp": 1699652599,
        "rating": 4
      },
      {
        "timestamp": 1699652600,
        "rating": 0
      },
      {
        "timestamp": 1699652601,
        "rating": 3
      }
    ]
  },
  ...
]
```

Basically, this file record each play. The rating is an integer between 0 and 5 (incl.), `0` meaning that you didn't enter a rating. The final rating is the mean of all the plays (`3.5` in the example above). You don't have to edit this file.

If you didn't set a rating file, then your ratings are ignored.


## Using a remote server

Instead of a metadata file, you can specify an url to get the files and ratings from a remote server. See my project `vgsserver` for a temporary implementation of such a server.


## Random notes

This is the new project that replaces the `vgsplayer` written in Python.

Depends on `mplayer`.

Tested on Linux (Debian-based) and MacOS.

License: MIT


## Want to talk?

Contact me at bruno@boberle.com
