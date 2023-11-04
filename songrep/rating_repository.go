package songrep

type PlayedSong struct {
	Path  string `json:"path"`
	Plays []Play `json:"plays"`
}

type Play struct {
	Timestamp int `json:"timestamp"`
	Rating    int `json:"rating"`
}

type RatingRepository interface {
	Rating(song Song) (float32, bool)
	AddPlay(song Song, timestamp int, rating int)
}
