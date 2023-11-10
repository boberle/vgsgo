package player

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"vgsgo/songrep"
	"strings"
	"testing"
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

func TestPlayer_getArgs(t *testing.T) {
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
			assert.Equal(t, tt.want, p.getArgs(tt.song))
		})
	}
}
