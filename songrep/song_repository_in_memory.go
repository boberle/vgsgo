package songrep

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type parsedSongs struct {
	Path           string  `json:"path"`
	Title          string  `json:"title"`
	GameTitle      string  `json:"game_title"`
	DurationSec    float32 `json:"duration"`
	LoopStartMicro int     `json:"loop_start"`
	LoopEndMicro   int     `json:"loop_end"`
	Size           int     `json:"size"`
	Error          bool    `json:"error"`
}

type parseFileResult struct {
	absPath string
	songs   []parsedSongs
}

type InMemorySongRepository struct {
	Songs            []Song
	RatingRepository RatingRepository
}

func SongsFromFiles(files []string) []Song {
	parsed := parseFiles(files)
	return convertImportedSongs(parsed)
}

func parseFiles(files []string) []parseFileResult {
	parsed := make([]parseFileResult, 0, len(files))
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		songs := make([]parsedSongs, 0, 5000)
		fh, err := os.Open(file)
		if err != nil {
			log.Fatalln(err)
		}
		defer func() {
			err := fh.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}()
		parseFile(fh, &songs)

		var abs string
		if filepath.IsAbs(file) {
			abs = filepath.Dir(file)
		} else {
			abs, err = filepath.Abs(filepath.Join(cwd, filepath.Dir(file)))
			if err != nil {
				log.Fatalln(err)
			}
		}
		parsed = append(parsed, parseFileResult{
			absPath: abs,
			songs:   songs,
		})
	}
	return parsed
}

func parseFile(fh io.Reader, songs *[]parsedSongs) {
	content, err := io.ReadAll(fh)
	if err != nil {
		log.Fatalln(err)
	}

	imported := make([]parsedSongs, 0, 1000)
	err = json.Unmarshal(content, &imported)
	if err != nil {
		log.Fatalln(err)
	}
	*songs = append(*songs, imported...)
}

func convertImportedSongs(parsed []parseFileResult) []Song {
	songCount := countParsedSongs(parsed)
	songs := make([]Song, 0, songCount)
	games := make(map[string]*Game)

	for _, p := range parsed {
		for _, s := range p.songs {
			if s.Error || s.Size == 0 {
				continue
			}
			fullPath := filepath.Join(p.absPath, s.Path)
			if _, found := games[s.GameTitle]; !found {
				game := Game{Title: s.GameTitle}
				games[s.GameTitle] = &game
			}
			songs = append(songs, makeSongFromImported(s, games[s.GameTitle], fullPath))
		}
	}
	return songs
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		log.Fatalln(err)
		return false // because goland complains
	}
}

func countParsedSongs(parsed []parseFileResult) int {
	total := 0
	for _, p := range parsed {
		total += len(p.songs)
	}
	return total
}

func makeSongFromImported(parsed parsedSongs, game *Game, absPath string) Song {
	return Song{
		Title:          parsed.Title,
		Game:           game,
		DurationSec:    parsed.DurationSec,
		LoopStartMicro: parsed.LoopStartMicro,
		LoopEndMicro:   parsed.LoopEndMicro,
		Path:           parsed.Path,
		AbsPath:        absPath,
	}
}

func (r *InMemorySongRepository) GetRandomSong(filters Filters) (Song, bool) {
	l := len(r.Songs)
	indices := getShuffledIndices(l, time.Now().Unix())
	index, found := r.getFirstFilteredSong(filters, indices)
	if found {
		song := r.Songs[index]
		song.IsPlayed = true
		return song, true
	}
	return Song{}, false
}

func (r *InMemorySongRepository) getFirstFilteredSong(filters Filters, indices []int) (int, bool) {
	for _, index := range indices {
		song := r.Songs[index]
		if song.IsPlayed {
			continue
		}
		if filters.MinDurationSec > 0 && song.DurationSec < float32(filters.MinDurationSec) {
			continue
		}
		if filters.TitleContains != "" && !strings.Contains(song.Title, filters.TitleContains) {
			continue
		}
		if filters.GameTitleContains != "" && !strings.Contains(song.Game.Title, filters.GameTitleContains) {
			continue
		}
		rating, found := r.RatingRepository.Rating(song)
		if filters.OnlyHasRating && !found {
			continue
		}
		if filters.OnlyHasNoRating && found {
			continue
		}
		if found && rating < filters.MinRating {
			continue
		}
		return index, true
	}
	return 0, false
}

func getShuffledIndices(n int, seed int64) []int {
	rv := make([]int, n)
	for i := 0; i < n; i++ {
		rv[i] = i
	}

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(n, func(i, j int) {
		rv[i], rv[j] = rv[j], rv[i]
	})
	return rv
}
