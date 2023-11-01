package songrep

import (
	"encoding/json"
	"io"
	"log"
)

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
	WriteJSON(writer io.Writer)
}

type InMemoryRatingRepository struct {
	PlayedSongs []PlayedSong
}

func RatingsFromJSON(reader io.Reader) InMemoryRatingRepository {
	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalln(err)
	}

	playedSongs := make([]PlayedSong, 0)
	err = json.Unmarshal(content, &playedSongs)
	if err != nil {
		log.Fatalln(err)
	}

	return InMemoryRatingRepository{PlayedSongs: playedSongs}
}

func (r *InMemoryRatingRepository) getSongByPath(path string) (*PlayedSong, bool) {
	for i, song := range r.PlayedSongs {
		if song.Path == path {
			return &r.PlayedSongs[i], true
		}
	}
	return &PlayedSong{}, false
}

func (r *InMemoryRatingRepository) AddPlay(song Song, timestamp int, rating int) {
	if s, found := r.getSongByPath(song.Path); found {
		s.Plays = append(s.Plays, Play{timestamp, rating})
	} else {
		r.PlayedSongs = append(r.PlayedSongs, PlayedSong{Path: song.Path, Plays: []Play{{timestamp, rating}}})
	}
}

func (r *InMemoryRatingRepository) Rating(song Song) (float32, bool) {
	if s, found := r.getSongByPath(song.Path); found {
		total := 0
		count := 0
		for _, p := range s.Plays {
			if p.Rating > 0 {
				total += p.Rating
				count++
			}
		}
		if count > 0 {
			return float32(total) / float32(count), true
		} else {
			return .0, false
		}
	} else {
		return .0, false
	}
}

func (r *InMemoryRatingRepository) WriteJSON(writer io.Writer) {
	content, err := json.Marshal(r.PlayedSongs)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = writer.Write(content)
	if err != nil {
		log.Fatalln(err)
	}
}
