package songrep

type Game struct {
	Title string
}

type Song struct {
	Title          string
	Game           *Game
	DurationSec    float32
	LoopStartMicro int
	LoopEndMicro   int
	Path           string
	AbsPath        string
	IsPlayed       bool
}

type Filters struct {
	MinRating         float32
	OnlyHasRating     bool
	OnlyHasNoRating   bool
	MinDurationSec    int
	TitleContains     string
	GameTitleContains string
}

type SongRepository interface {
	GetRandomSong(filters Filters) (Song, bool)
	MarkAsPlayed(song Song) error
}

type SongNotFound struct {
	song Song
}

func (s SongNotFound) Error() string {
	return s.song.Path
}
