package player

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"vgsgo/songrep"
)

func TestPlayer_Rate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RatingAction
	}{
		{"nothing", "\n", RatingAction{0, false, false}},
		{"rating", "2\n", RatingAction{2, false, false}},
		{"rating resume", "2r\n", RatingAction{2, true, false}},
		{"rating resume 2", "2 r\n", RatingAction{2, true, false}},
		{"resume", "2 r\n", RatingAction{2, true, false}},
		{"rating quit", "2 q\n", RatingAction{2, false, true}},
		{"rating quit 2", "2q\n", RatingAction{2, false, true}},
		{"quit", "q\n", RatingAction{0, false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Player{
				Input:  strings.NewReader(tt.input),
				Output: bytes.NewBuffer([]byte{}),
			}
			got := p.Rate()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlayer_getArgsWithMaxPlays(t *testing.T) {
	tests := []struct {
		name    string
		maxPlay int
		song    songrep.Song
		want    []string
	}{
		{"one play", 1, songrep.Song{AbsPath: "/foo/bar.brstm"}, []string{"mplayer", "/foo/bar.brstm"}},
		{"one play", 1, songrep.Song{AbsPath: "/foo/bar.brstm", LoopEndMicro: 1234}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "0.001234"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm"}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-loop", "1"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm", LoopStartMicro: 1234567}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-ss", "1.234567", "-loop", "1"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm", LoopStartMicro: 1234567, LoopEndMicro: 7654321}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "7.654321", "/foo/bar.brstm", "-ss", "1.234567", "-endpos", "7.654321", "-loop", "1"}},
		{"infinite loop", 0, songrep.Song{AbsPath: "/foo/bar.brstm", LoopStartMicro: 1234567, LoopEndMicro: 7654321}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "7.654321", "/foo/bar.brstm", "-ss", "1.234567", "-endpos", "7.654321", "-loop", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Player{
				Cmd:     "mplayer",
				MaxPlay: tt.maxPlay,
			}
			assert.Equal(t, tt.want, p.getArgsWithMaxPlays(tt.song))
		})
	}
}

func TestPlayer_getArgsWithMaxPlayTime(t *testing.T) {
	tests := []struct {
		name        string
		maxPlayTime int
		song        songrep.Song
		want        []string
	}{
		{"one play, song is longer", 5, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 10}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "5"}},
		{"3 complete plays", 10, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 3}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "/foo/bar.brstm"}},
		{"1 complete play + 1 incomplete", 6, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 3}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-endpos", "3"}},
		{"2 complete plays + 1 incomplete", 16, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 6}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "/foo/bar.brstm", "-endpos", "4"}},
		{"3 complete plays + 1 incomplete", 22, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 6}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "/foo/bar.brstm", "/foo/bar.brstm", "-endpos", "4"}},
		{"3 complete plays + 0 (too short), with loop start", 14, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 5, LoopStartMicro: 1500000}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-ss", "1.500000", "/foo/bar.brstm", "-ss", "1.500000"}},
		{"3 complete plays + 1 incomplete, with loop start", 32, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 10, LoopStartMicro: 1500000}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-ss", "1.500000", "/foo/bar.brstm", "-ss", "1.500000", "/foo/bar.brstm", "-ss", "1.500000", "-endpos", "6"}},
		{"3 complete plays + 1 incomplete, with loop start and end", 26, songrep.Song{AbsPath: "/foo/bar.brstm", DurationSec: 10, LoopStartMicro: 1500000, LoopEndMicro: 8000000}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "8.000000", "/foo/bar.brstm", "-ss", "1.500000", "-endpos", "8.000000", "/foo/bar.brstm", "-ss", "1.500000", "-endpos", "8.000000", "/foo/bar.brstm", "-ss", "1.500000", "-endpos", "6"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Player{
				Cmd:            "mplayer",
				MaxPlayTimeSec: tt.maxPlayTime,
			}
			assert.Equal(t, tt.want, p.getArgsWithMaxPlayTime(tt.song))
		})
	}
}
