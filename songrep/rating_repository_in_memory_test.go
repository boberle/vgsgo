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
		songs      []PlayedSong
		wantRating float32
		wantFound  bool
	}{
		{"0 play", []PlayedSong{}, .0, false},
		{"1 play", []PlayedSong{{"path", []Play{{0, 2}}}}, 2.0, true},
		{"2 plays", []PlayedSong{{"path", []Play{{0, 1}, {0, 2}}}}, 1.5, true},
		{"3 plays", []PlayedSong{{"path", []Play{{0, 1}, {0, 2}, {0, 3}}}}, 2.0, true},
		{"2/3 plays", []PlayedSong{{"path", []Play{{0, 1}, {0, 1}}}, {"path2", []Play{{0, 2}}}}, 1.0, true},
		{"3, one with no rating", []PlayedSong{{"path", []Play{{0, 1}, {0, 0}, {0, 1}}}}, 1.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemoryRatingRepository{
				PlayedSongs: tt.songs,
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
		songs []PlayedSong
		want  string
	}{
		{"0 play", []PlayedSong{}, "[]"},
		{"1 play", []PlayedSong{{"path", []Play{{123, 2}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":123,\"rating\":2}]}]"},
		{"2 plays, 1 song", []PlayedSong{{"path", []Play{{456, 1}, {789, 2}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":456,\"rating\":1},{\"timestamp\":789,\"rating\":2}]}]"},
		{"2 plays, 2 song", []PlayedSong{{"path", []Play{{123, 1}}}, {"path2", []Play{{456, 2}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":123,\"rating\":1}]},{\"path\":\"path2\",\"plays\":[{\"timestamp\":456,\"rating\":2}]}]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemoryRatingRepository{
				PlayedSongs: tt.songs,
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
		want InMemoryRatingRepository
		data string
	}{
		{"0 play", InMemoryRatingRepository{PlayedSongs: []PlayedSong{}}, "[]"},
		{"1 play", InMemoryRatingRepository{PlayedSongs: []PlayedSong{{"path", []Play{{123, 2}}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":123,\"rating\":2}]}]"},
		{"2 plays, 1 song", InMemoryRatingRepository{PlayedSongs: []PlayedSong{{"path", []Play{{456, 1}, {789, 2}}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":456,\"rating\":1},{\"timestamp\":789,\"rating\":2}]}]"},
		{"2 plays, 2 song", InMemoryRatingRepository{PlayedSongs: []PlayedSong{{"path", []Play{{123, 1}}}, {"path2", []Play{{456, 2}}}}}, "[{\"path\":\"path\",\"plays\":[{\"timestamp\":123,\"rating\":1}]},{\"path\":\"path2\",\"plays\":[{\"timestamp\":456,\"rating\":2}]}]"},
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
	}
	tests := []struct {
		name string
		args []args
		want []PlayedSong
	}{
		{"1 call", []args{{Song{Path: "foo"}, 1, 1}}, []PlayedSong{{"foo", []Play{{1, 1}}}}},
		{"2 calls", []args{{Song{Path: "foo"}, 1, 1}, {Song{Path: "bar"}, 2, 2}}, []PlayedSong{{"foo", []Play{{1, 1}}}, {"bar", []Play{{2, 2}}}}},
		{"3 calls", []args{{Song{Path: "foo"}, 1, 1}, {Song{Path: "bar"}, 2, 2}, {Song{Path: "foo"}, 3, 3}}, []PlayedSong{{"foo", []Play{{1, 1}, {3, 3}}}, {"bar", []Play{{2, 2}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &InMemoryRatingRepository{}
			for _, args := range tt.args {
				r.AddPlay(args.song, args.timestamp, args.rating)
			}
			assert.Equal(t, tt.want, r.PlayedSongs)
		})
	}
}
