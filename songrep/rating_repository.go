package songrep

import (
	"encoding/json"
	"io"
	"log"
)

type Play struct {
	Path      string `json:"path"`
	Timestamp int    `json:"timestamp"`
	Rating    int    `json:"rating"`
	NoRating  bool   `json:"no_rating"`
}

type RatingRepository interface {
	Rating(song Song) (float32, bool)
	AddPlay(song Song, timestamp int, rating int, noRating bool)
	WriteJSON(writer io.Writer)
}

type InMemoryRatingRepository struct {
	Plays []Play
}

func RatingsFromJSON(reader io.Reader) InMemoryRatingRepository {
	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalln(err)
	}

	plays := make([]Play, 0)
	err = json.Unmarshal(content, &plays)
	if err != nil {
		log.Fatalln(err)
	}

	return InMemoryRatingRepository{Plays: plays}
}

func (r *InMemoryRatingRepository) AddPlay(song Song, timestamp int, rating int, noRating bool) {
	r.Plays = append(r.Plays, Play{song.Path, timestamp, rating, noRating})
}

func (r *InMemoryRatingRepository) Rating(song Song) (float32, bool) {
	total := 0
	count := 0
	for _, p := range r.Plays {
		if p.Path == song.Path && p.NoRating == false {
			total += p.Rating
			count++
		}
	}
	if count > 0 {
		return float32(total) / float32(count), true
	} else {
		return .0, false
	}
}

func (r *InMemoryRatingRepository) WriteJSON(writer io.Writer) {
	content, err := json.Marshal(r.Plays)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = writer.Write(content)
	if err != nil {
		log.Fatalln(err)
	}
}
