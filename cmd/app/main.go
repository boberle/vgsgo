package main

import (
	"flag"
	"fmt"
	playerpck "gosmash/player"
	"gosmash/songrep"
	"log"
	"os"
	"time"
)

func main() {

	args := getArgs()

	var ratingRep songrep.RatingRepository
	if _, err := os.Stat(args.ratings); err == nil {
		fh, err := os.Open(args.ratings)
		if err != nil {
			log.Fatal(err)
		}
		rep := songrep.RatingsFromJSON(fh)
		rep.File = args.ratings
		_ = fh.Close()
		ratingRep = &rep
	} else {
		ratingRep = &songrep.InMemoryRatingRepository{File: args.ratings}
	}

	songRep := songrep.InMemorySongRepository{
		Songs:            songrep.SongsFromFiles(args.dbFiles),
		RatingRepository: ratingRep,
	}

	player := playerpck.Player{
		Cmd:     "/usr/bin/mplayer",
		Input:   os.Stdin,
		Output:  os.Stdout,
		MaxPlay: args.maxPlays,
	}

	filters := songrep.Filters{
		MinRating:         float32(args.minRating),
		OnlyHasRating:     args.onlyHasRating,
		OnlyHasNoRating:   args.onlyHasNoRating,
		MinDurationSec:    args.minDurationSec,
		TitleContains:     args.titleContains,
		GameTitleContains: args.gameTitleContains,
	}

	run(&songRep, ratingRep, player, filters, args.ratings)

}

func run(songRep songrep.SongRepository, ratingRep songrep.RatingRepository, player playerpck.Player, filters songrep.Filters, ratingFile string) {
	for {
		song, found := songRep.GetRandomSong(filters)
		if !found {
			fmt.Println("no more song")
			return
		}
		player.Play(song)
		err := songRep.MarkAsPlayed(song)
		if err != nil {
			log.Println(err)
		}

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
		_, _ = fmt.Fprintln(os.Stderr, "You must provide one or more db files")
		flag.Usage()
		os.Exit(1)
	}

	if args.ratings == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Argument -rating-file is required")
		flag.Usage()
		os.Exit(1)
	}

	args.dbFiles = flag.Args()
	return args
}
