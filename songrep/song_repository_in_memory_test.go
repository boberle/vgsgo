package songrep

import (
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func Test_getShuffledIndices(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []int
	}{
		{"n is 0", 0, []int{}},
		{"n is 1", 1, []int{0}},
		{"n is 2", 2, []int{0, 1}},
		{"n is 5", 5, []int{4, 1, 3, 0, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShuffledIndices(tt.n, 123)
			assert.Equalf(t, tt.want, got, "getShuffledIndices(%v, %v)", tt.n, 123)
		})
	}
}

func TestInMemorySongRepository_getFirstFilteredSong(t *testing.T) {
	game1 := Game{Title: "foo abc"}
	game2 := Game{Title: "foo def"}
	songs := []Song{
		{Title: "bar ghi", Game: &game1, DurationSec: 10, Path: "foo"}, // "shuffled index": 2
		{Title: "bar jkl", Game: &game1, DurationSec: 20, Path: "bar"}, // "shuffled index": 1
		{Title: "bar mno", Game: &game2, DurationSec: 30, Path: "baz"}, // "shuffled index": 0
		{Title: "bar pqr", Game: &game2, DurationSec: 40, Path: "biz"}, // "shuffled index": 3
	}
	ratings := InMemoryRatingRepository{
		PlayedSongs: []PlayedSong{
			{"foo", []Play{{Rating: 5}, {Rating: 0}, {Rating: 3}}},
			{"biz", []Play{{Rating: 2}}},
			{"baz", []Play{{Rating: 1}}},
		},
	}
	tests := []struct {
		name      string
		filters   Filters
		wantIndex int
		wantFound bool
	}{
		{"no filter", Filters{}, 2, true},
		{"duration >= 10", Filters{MinDurationSec: 10}, 2, true},
		{"duration >= 20", Filters{MinDurationSec: 20}, 2, true},
		{"duration >= 30", Filters{MinDurationSec: 30}, 2, true},
		{"duration >= 40", Filters{MinDurationSec: 40}, 3, true},
		{"duration >= 50", Filters{MinDurationSec: 50}, 0, false},
		{"song title 'bar'", Filters{TitleContains: "bar"}, 2, true},
		{"song title 'jkl'", Filters{TitleContains: "jkl"}, 1, true},
		{"song title 'not found'", Filters{TitleContains: "not found"}, 0, false},
		{"game title 'foo'", Filters{GameTitleContains: "foo"}, 2, true},
		{"game title 'abc'", Filters{GameTitleContains: "abc"}, 1, true},
		{"game title 'not found'", Filters{GameTitleContains: "not found"}, 0, false},
		{"3 filters", Filters{MinDurationSec: 20, TitleContains: "bar", GameTitleContains: "abc"}, 1, true},
		{"3 filters, not found", Filters{MinDurationSec: 30, TitleContains: "bar", GameTitleContains: "abc"}, 0, false},
		// ratings
		{"rating >= 4", Filters{MinRating: 4, OnlyHasRating: true}, 0, true},
		{"rating >= 4 or no rating", Filters{MinRating: 4}, 1, true},
		{"no rating", Filters{OnlyHasNoRating: true}, 1, true},
		{"rating", Filters{OnlyHasRating: true}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemorySongRepository{
				Songs:            songs,
				RatingRepository: ratings,
			}
			gotIndex, gotFound := r.getFirstFilteredSong(tt.filters, []int{2, 1, 0, 3})
			assert.Equalf(t, tt.wantIndex, gotIndex, "GetRandomSong(%v)", tt.filters)
			assert.Equalf(t, tt.wantFound, gotFound, "GetRandomSong(%v)", tt.filters)
		})
	}
}

func TestInMemorySongRepository_GetRandomSong(t *testing.T) {
	game1 := Game{Title: "foo abc"}
	songs := []Song{
		{Title: "bar def", Game: &game1, DurationSec: 10},
		{Title: "bar ghi", Game: &game1, DurationSec: 20},
	}
	tests := []struct {
		name      string
		filters   Filters
		wantFound bool
	}{
		{"no filter", Filters{}, true},
		{"not found", Filters{MinDurationSec: 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := InMemorySongRepository{
				Songs:            songs,
				RatingRepository: InMemoryRatingRepository{},
			}
			gotSong, gotFound := r.GetRandomSong(tt.filters)
			if gotFound && gotSong.Title != "bar def" && gotSong.Title != "bar ghi" {
				t.Errorf("song not found")
			}
			assert.Equalf(t, tt.wantFound, gotFound, "GetRandomSong(%v)", tt.filters)
			assert.Equalf(t, tt.wantFound, gotSong.IsPlayed, "GetRandomSong(%v)", tt.filters)
		})
	}
}

func Test_parseFile(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []parsedSongs
	}{
		{"empty", "[]", []parsedSongs{}},
		{
			"1 entry",
			"[{\"path\":\"path1\",\"timestamp\":123,\"title\":\"abc\",\"game_title\":\"ABC\",\"duration\":1.23,\"loop_start\":2,\"loop_end\":3,\"size\":4,\"error\":false}]",
			[]parsedSongs{{"path1", "abc", "ABC", 1.23, 2, 3, 4, false}},
		},
		{
			"2 entries",
			"[{\"path\":\"path1\",\"timestamp\":123,\"title\":\"abc\",\"game_title\":\"ABC\",\"duration\":1,\"loop_start\":2,\"loop_end\":3,\"size\":4,\"error\":false},{\"path\":\"path2\",\"timestamp\":456,\"title\":\"def\",\"game_title\":\"DEF\",\"duration\":5,\"loop_start\":6,\"loop_end\":7,\"size\":8,\"error\":true}]",
			[]parsedSongs{
				{"path1", "abc", "ABC", 1, 2, 3, 4, false},
				{"path2", "def", "DEF", 5, 6, 7, 8, true},
			},
		},
		{
			"no title",
			"[{\"path\":\"path1\",\"timestamp\":123,\"title\":null,\"game_title\":null,\"duration\":1,\"loop_start\":2,\"loop_end\":3,\"size\":4,\"error\":false}]",
			[]parsedSongs{{"path1", "", "", 1, 2, 3, 4, false}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.data)
			got := make([]parsedSongs, 0)
			parseFile(r, &got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSongsFromFiles(t *testing.T) {
	files := []string{"testdata/songs.json", "testdata/abc/songs.json"}
	game1 := Game{Title: "ABC"}
	game2 := Game{Title: "DEF"}
	cwd, _ := os.Getwd()
	cwd += "/"
	want := []Song{
		{"abc", &game1, 1, 2, 3, "hello/foo.brstm", cwd + "testdata/hello/foo.brstm", false},
		{"def", &game2, 4, 5, 6, "bar.brstm", cwd + "testdata/bar.brstm", false},
		{"abc", &game1, 1.23, 2, 3, "hello/foo.brstm", cwd + "testdata/abc/hello/foo.brstm", false},
		{"def", &game2, 4, 5, 6, "bar.brstm", cwd + "testdata/abc/bar.brstm", false},
	}
	got := SongsFromFiles(files)
	td.Cmp(t, got, want)
}
