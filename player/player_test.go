package player

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gosmash/songrep"
	"strings"
	"testing"
)

func TestPlayer_Rate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RatingAction
	}{
		{"nothing", "\n", RatingAction{0, true, false, false}},
		{"rating", "2\n", RatingAction{2, false, false, false}},
		{"rating resume", "2r\n", RatingAction{2, false, true, false}},
		{"rating resume 2", "2 r\n", RatingAction{2, false, true, false}},
		{"resume", "2 r\n", RatingAction{2, false, true, false}},
		{"rating quit", "2 q\n", RatingAction{2, false, false, true}},
		{"rating quit 2", "2q\n", RatingAction{2, false, false, true}},
		{"quit", "q\n", RatingAction{0, true, false, true}},
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

func TestPlayer_getArgs(t *testing.T) {
	tests := []struct {
		name    string
		maxPlay int
		song    songrep.Song
		want    []string
	}{
		{"one play", 1, songrep.Song{AbsPath: "/foo/bar.brstm"}, []string{"mplayer", "/foo/bar.brstm"}},
		{"one play", 1, songrep.Song{AbsPath: "/foo/bar.brstm", EndLoopMilli: 1000}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "1000"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm"}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-loop", "1"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm", StartLoopMilli: 200}, []string{"mplayer", "/foo/bar.brstm", "/foo/bar.brstm", "-ss", "200", "-loop", "1"}},
		{"2 plays", 2, songrep.Song{AbsPath: "/foo/bar.brstm", StartLoopMilli: 200, EndLoopMilli: 1000}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "1000", "/foo/bar.brstm", "-ss", "200", "-endpos", "1000", "-loop", "1"}},
		{"infinite loop", 0, songrep.Song{AbsPath: "/foo/bar.brstm", StartLoopMilli: 200, EndLoopMilli: 1000}, []string{"mplayer", "/foo/bar.brstm", "-endpos", "1000", "/foo/bar.brstm", "-ss", "200", "-endpos", "1000", "-loop", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Player{
				Cmd:     "mplayer",
				MaxPlay: tt.maxPlay,
			}
			assert.Equal(t, tt.want, p.getArgs(tt.song))
		})
	}
}
