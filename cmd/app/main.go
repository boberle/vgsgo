package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"time"
	playerpck "vgsgo/player"
	"vgsgo/songrep"
)

func main() {

	args := getArgs()

	player := playerpck.Player{
		Cmd:            "/usr/bin/mplayer",
		Input:          os.Stdin,
		Output:         os.Stdout,
		MaxPlays:       args.maxPlays,
		MaxPlayTimeSec: args.maxPlayTime,
	}

	filters := songrep.Filters{
		MinRating:         float32(args.minRating),
		OnlyHasRating:     args.onlyHasRating,
		OnlyHasNoRating:   args.onlyHasNoRating,
		MinDurationSec:    args.minDurationSec,
		TitleContains:     args.titleContains,
		GameTitleContains: args.gameTitleContains,
	}

	conf := getConfiguration(args)
	run(conf.songRep, conf.ratingRep, player, filters, args.ratings)

}

func run(songRep songrep.SongRepository, ratingRep songrep.RatingRepository, player playerpck.Player, filters songrep.Filters, ratingFile string) {
	for {
		song, found := songRep.GetRandomSong(filters)
		if !found {
			fmt.Println("no more song")
			return
		}
		player.Play(song)

		actions := player.Rate()
		rating := actions.Value
		ratingRep.AddPlay(song, int(time.Now().Unix()), rating)
		if actions.Resume {
			player.PlayIndefinitely(song)
		}

		if actions.Quit {
			return
		}
	}
}

type Arguments struct {
	dbFiles           []string
	ratings           string
	maxPlays          int
	maxPlayTime       int
	continuousPlay    bool
	playLast          bool
	minRating         float64
	onlyHasRating     bool
	onlyHasNoRating   bool
	minDurationSec    int
	titleContains     string
	gameTitleContains string
}

func getArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.ratings, "rating-file", "", "json file where ratings are store")
	flag.IntVar(&args.maxPlays, "max-plays", 0, "maximum number of plays (default is 0, infinity)")
	flag.IntVar(&args.maxPlayTime, "max-play-time", 0, "maximum time to play (default is 0, infinity)")
	flag.BoolVar(&args.continuousPlay, "continuous", false, "don't stop to ask rating")
	flag.BoolVar(&args.playLast, "play-last", false, "don't shuffle songs, play the last ones")
	flag.Float64Var(&args.minRating, "min-rating", 0, "minimum rating. Add --only-has-rating to limit to songs that have ratings")
	flag.BoolVar(&args.onlyHasRating, "only-has-rating", false, "limit to songs that have a rating")
	flag.BoolVar(&args.onlyHasNoRating, "only-has-no-rating", false, "limit to songs that don't have a rating")
	flag.IntVar(&args.minDurationSec, "min-duration", 0, "minimum duration")
	flag.StringVar(&args.titleContains, "title", "", "limit to song with a title that contains the string")
	flag.StringVar(&args.gameTitleContains, "game-title", "", "limit to song with a game title that contains the string")

	flag.Parse()

	if flag.NArg() == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "You must provide one or more db files, or an url to a server")
		flag.Usage()
		os.Exit(1)
	}

	if args.maxPlays != 0 && args.maxPlayTime != 0 {
		fmt.Println("You can't use -max-plays and -max-play-time at the same time")
		os.Exit(1)
	}

	args.dbFiles = flag.Args()
	return args
}

type AppConfiguration struct {
	ratingRep songrep.RatingRepository
	songRep   songrep.SongRepository
}

func getConfiguration(args Arguments) AppConfiguration {
	if len(args.dbFiles) == 1 && strings.HasPrefix(args.dbFiles[0], "http") {
		return getRemoteConfiguration(args)
	} else {
		return getLocalConfiguration(args)
	}
}

func getLocalConfiguration(args Arguments) AppConfiguration {
	var ratingRep songrep.InMemoryRatingRepository
	if _, err := os.Stat(args.ratings); err == nil {
		fh, err := os.Open(args.ratings)
		if err != nil {
			log.Fatal(err)
		}
		rep := songrep.RatingsFromJSON(fh)
		rep.File = args.ratings
		_ = fh.Close()
		ratingRep = rep
	} else {
		ratingRep = songrep.InMemoryRatingRepository{File: args.ratings}
	}

	songRep := songrep.InMemorySongRepository{
		Songs:            songrep.SongsFromFiles(args.dbFiles),
		RatingRepository: ratingRep,
	}

	return AppConfiguration{
		ratingRep: &ratingRep,
		songRep:   &songRep,
	}
}

func getRemoteConfiguration(args Arguments) AppConfiguration {
	var username, password string
	s := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your username: ")
	if s.Scan() {
		username = s.Text()
	}

	fmt.Print("Enter your password: ")
	bytePasswd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalln(err)
	}
	password = string(bytePasswd)

	ratingRep := songrep.RemoteRatingRepository{
		ServerBaseUrl: args.dbFiles[0],
		Username:      username,
		Password:      password,
	}

	songRep := songrep.RemoteSongRepository{
		ServerBaseUrl: args.dbFiles[0],
		Username:      username,
		Password:      password,
		SongDir:       "/tmp/vgsgo/songs",
	}

	return AppConfiguration{
		ratingRep: &ratingRep,
		songRep:   &songRep,
	}
}
