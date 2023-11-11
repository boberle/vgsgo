package player

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"vgsgo/songrep"
)

type Player struct {
	Cmd            string
	Input          io.Reader
	Output         io.Writer
	MaxPlay        int
	MaxPlayTimeSec int
}

type RatingAction struct {
	Value  int
	Resume bool
	Quit   bool
}

func (p Player) Play(song songrep.Song) {
	var args []string
	if p.MaxPlayTimeSec != 0 {
		args = p.getArgsWithMaxPlayTime(song)
	} else {
		args = p.getArgsWithMaxPlays(song)
	}
	p.exec(args)
}

func (p Player) PlayIndefinitely(song songrep.Song) {
	player := p
	player.MaxPlay = 0
	args := player.getArgsWithMaxPlays(song)
	p.exec(args)
}

func (p Player) Rate() RatingAction {
	for {
		_, err := fmt.Fprintf(p.Output, "What ([<int>] [r] [q])? ")
		if err != nil {
			log.Fatalln(err)
		}

		scanner := bufio.NewScanner(p.Input)
		if scanner.Scan() {
			pat, err := regexp.Compile(`^(?i:\s*([12345]+)?\s*(r)?\s*(q)?)$`)
			if err != nil {
				log.Fatal(err)
			}

			cmd := strings.TrimSpace(scanner.Text())
			matches := pat.FindStringSubmatch(cmd)
			if matches == nil {
				continue
			}

			value := 0
			var resume, quit bool
			if matches[1] != "" {
				value, _ = strconv.Atoi(matches[1])
			}

			if matches[2] != "" {
				resume = true
			}

			if matches[3] != "" {
				quit = true
			}

			return RatingAction{
				Value:  value,
				Resume: resume,
				Quit:   quit,
			}
		}
	}
}

func (p Player) getArgsWithMaxPlays(song songrep.Song) []string {
	args := make([]string, 0, 10)
	args = append(args, p.Cmd)

	// first run (start from 0)
	args = append(args, song.AbsPath)
	if song.LoopEndMicro != 0 {
		args = append(args, "-endpos")
		args = append(args, fmt.Sprintf("%f", float32(song.LoopEndMicro)/1000000.0))
	}

	// other run (start from startLoop)
	if p.MaxPlay != 1 {
		args = append(args, song.AbsPath)
		if song.LoopStartMicro != 0 {
			args = append(args, "-ss")
			args = append(args, fmt.Sprintf("%f", float32(song.LoopStartMicro)/1000000.0))
		}
		if song.LoopEndMicro != 0 {
			args = append(args, "-endpos")
			args = append(args, fmt.Sprintf("%f", float32(song.LoopEndMicro)/1000000.0))
		}
		args = append(args, "-loop")
		if p.MaxPlay == 0 {
			args = append(args, "0")
		} else {
			args = append(args, strconv.Itoa(p.MaxPlay-1))
		}
	}
	return args
}

func (p Player) getArgsWithMaxPlayTime(song songrep.Song) []string {
	if p.MaxPlayTimeSec == 0 {
		panic("MaxPlayTimeSec can't be 0")
	}

	args := make([]string, 0, 10)
	args = append(args, p.Cmd)

	// first run (start from 0)
	args = append(args, song.AbsPath)
	loopEndSec := float32(song.LoopEndMicro) / 1000000.0
	var elapsed = song.DurationSec
	if song.DurationSec > float32(p.MaxPlayTimeSec) {
		args = append(args, "-endpos")
		args = append(args, fmt.Sprintf("%d", p.MaxPlayTimeSec))
		elapsed = float32(p.MaxPlayTimeSec)
	} else if loopEndSec != 0.0 {
		args = append(args, "-endpos")
		args = append(args, fmt.Sprintf("%f", loopEndSec))
		elapsed = loopEndSec
	}

	// other run (start from startLoop)
	loopStartSec := float32(song.LoopStartMicro) / 1000000.0
	var playTime float32
	if loopEndSec == 0 {
		playTime = song.DurationSec - loopStartSec
	} else {
		playTime = loopEndSec - loopStartSec
	}
	for elapsed+playTime < float32(p.MaxPlayTimeSec) {
		args = append(args, song.AbsPath)
		if song.LoopStartMicro != 0 {
			args = append(args, "-ss")
			args = append(args, fmt.Sprintf("%f", loopStartSec))
		}
		if song.LoopEndMicro != 0 {
			args = append(args, "-endpos")
			args = append(args, fmt.Sprintf("%f", loopEndSec))
		}
		elapsed += playTime
	}

	// last run
	if float32(p.MaxPlayTimeSec)-elapsed > 2.0 {
		args = append(args, song.AbsPath)
		if song.LoopStartMicro != 0 {
			args = append(args, "-ss")
			args = append(args, fmt.Sprintf("%f", loopStartSec))
		}
		args = append(args, "-endpos")
		args = append(args, fmt.Sprintf("%.0f", float32(p.MaxPlayTimeSec)-elapsed+loopStartSec))
	}

	return args
}

func (p Player) exec(args []string) {
	proc, err := os.StartProcess(
		p.Cmd,
		args,
		&os.ProcAttr{
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = proc.Wait()
	if err != nil {
		log.Fatalln(err)
	}
}
