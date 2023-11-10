package player

import (
	"bufio"
	"fmt"
	"vgsgo/songrep"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Player struct {
	Cmd     string
	Input   io.Reader
	Output  io.Writer
	MaxPlay int
}

type RatingAction struct {
	Value  int
	Resume bool
	Quit   bool
}

func (p Player) Play(song songrep.Song) {
	args := p.getArgs(song)
	p.exec(args)
}

func (p Player) PlayIndefinitely(song songrep.Song) {
	player := p
	player.MaxPlay = 0
	args := player.getArgs(song)
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

func (p Player) getArgs(song songrep.Song) []string {
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
