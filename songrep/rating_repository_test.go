package songrep

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestInMemoryRatingRepository_Rating(t *testing.T) {
	tests := []struct {
		name       string
		plays      []Play
		wantRating float32
		wantFound  bool
	}{
		{"0 play", []Play{}, .0, false},
		{"1 play", []Play{{"path", 0, 2, false}}, 2.0, true},
		{"2 plays", []Play{{"path", 0, 1, false}, {"path", 0, 2, false}}, 1.5, true},
		{"3 plays", []Play{{"path", 0, 1, false}, {"path", 0, 2, false}, {"path", 0, 3, false}}, 2.0, true},
		{"2/3 plays", []Play{{"path", 0, 1, false}, {"path2", 0, 2, false}, {"path", 0, 1, false}}, 1.0, true},
		{"3, one with no rating", []Play{{"path", 0, 1, false}, {"path", 0, 2, true}, {"path", 0, 1, false}}, 1.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemoryRatingRepository{
				Plays: tt.plays,
			}
			if gotRating, gotFound := r.Rating(Song{Path: "path"}); gotRating != tt.wantRating || gotFound != tt.wantFound {
				t.Errorf("Rating() = %v, want %v", gotRating, tt.wantRating)
				t.Errorf("Rating() = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestInMemoryRatingRepository_writeJSON(t *testing.T) {
	tests := []struct {
		name  string
		plays []Play
		want  string
	}{
		{"0 play", []Play{}, "[]"},
		{"1 play", []Play{{"path", 123, 2, false}}, "[{\"path\":\"path\",\"timestamp\":123,\"rating\":2,\"no_rating\":false}]"},
		{"2 plays", []Play{{"path", 456, 1, false}, {"path2", 789, 2, false}}, "[{\"path\":\"path\",\"timestamp\":456,\"rating\":1,\"no_rating\":false},{\"path\":\"path2\",\"timestamp\":789,\"rating\":2,\"no_rating\":false}]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemoryRatingRepository{
				Plays: tt.plays,
			}
			buf := bytes.NewBuffer([]byte{})
			r.WriteJSON(buf)
			got := buf.String()
			if got != tt.want {
				t.Errorf("Rating() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRatingsFromJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
		want InMemoryRatingRepository
	}{
		{"0 play", "[]", InMemoryRatingRepository{[]Play{}}},
		{"1 play", "[{\"path\":\"path\",\"timestamp\":123,\"rating\":2}]", InMemoryRatingRepository{[]Play{{"path", 123, 2, false}}}},
		{"2 plays", "[{\"path\":\"path\",\"timestamp\":456,\"rating\":1},{\"path\":\"path2\",\"timestamp\":789,\"rating\":2}]", InMemoryRatingRepository{[]Play{{"path", 456, 1, false}, {"path2", 789, 2, false}}}},
	}
	for _, tt := range tests {
		r := strings.NewReader(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RatingsFromJSON(r), "RatingsFromJSON(%v)", tt.data)
		})
	}
}

func TestInMemoryRatingRepository_AddPlay(t *testing.T) {
	type args struct {
		song      Song
		timestamp int
		rating    int
		noRating  bool
	}
	tests := []struct {
		name string
		args []args
		want int
	}{
		{"one call", []args{{Song{Path: "foo"}, 1, 1, false}}, 1},
		{"two calls", []args{{Song{Path: "foo"}, 1, 1, false}, {Song{Path: "bar"}, 2, 2, false}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InMemoryRatingRepository{}
			for _, args := range tt.args {
				r.AddPlay(args.song, args.timestamp, args.rating, args.noRating)
			}
			assert.Equal(t, tt.want, len(r.Plays))
		})
	}
}
