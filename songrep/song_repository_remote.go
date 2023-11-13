package songrep

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type RemoteSongRepository struct {
	ServerBaseUrl string
	SongDir       string
	Username      string
	Password      string
}

func (r *RemoteSongRepository) GetRandomSong(filters Filters) (Song, bool) {
	// TODO: filters
	randomSongUrl := r.ServerBaseUrl + "/api/songs/random/"
	song, id, found := downloadSongMetadata(randomSongUrl, r.SongDir, r.Username, r.Password, filters)
	if !found {
		return Song{}, false
	}
	songFileDir := filepath.Dir(song.AbsPath)
	err := os.MkdirAll(songFileDir, 0700)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: don't download if already there
	songFileUrl := r.ServerBaseUrl + "/api/songs/" + id + "/file/"
	err = downloadSongFile(songFileUrl, song.AbsPath, r.Username, r.Password)
	if err != nil {
		log.Fatalln(err)
	}

	return song, true
}

func setAuthHeader(req *http.Request, username, password string) {
	authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("Authorization", "Basic "+authHeader)
}

func downloadSongMetadata(url, songDir, username, password string, filters Filters) (Song, string, bool) {
	req, _ := http.NewRequest("GET", url, nil)
	setAuthHeader(req, username, password)
	client := &http.Client{}

	values := req.URL.Query()
	if filters.MinRating > 0. {
		values.Set("min_rating", strconv.Itoa(int(filters.MinRating)))
	}
	if filters.MinDurationSec > 0. {
		values.Set("min_duration", strconv.Itoa(filters.MinDurationSec))
	}
	if filters.OnlyHasRating {
		values.Set("only_has_rating", "true")
	}
	if filters.OnlyHasNoRating {
		values.Set("only_has_no_rating", "true")
	}
	if len(filters.TitleContains) > 0 {
		values.Set("title_contains", filters.TitleContains)
	}
	if len(filters.GameTitleContains) > 0 {
		values.Set("game_title_contains", filters.GameTitleContains)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 404 {
		return Song{}, "", false
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Return code is not 200: %d\n", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if !json.Valid(body) {
		log.Fatalln("response json is not valid")
	}

	type songResponse struct {
		Id        string  `json:"id"`
		Title     string  `json:"title"`
		GameTitle string  `json:"game_title"`
		Duration  float32 `json:"duration"`
		LoopStart int     `json:"loop_start"`
		LoopEnd   int     `json:"loop_end"`
		Path      string  `json:"path"`
	}

	var songResp songResponse
	err = json.Unmarshal(body, &songResp)
	if err != nil {
		log.Fatalln(err)
	}

	song := Song{
		Title: songResp.Title,
		Game: &Game{
			Title: songResp.GameTitle,
		},
		DurationSec:    songResp.Duration,
		LoopStartMicro: songResp.LoopStart,
		LoopEndMicro:   songResp.LoopEnd,
		Path:           songResp.Path,
		AbsPath:        filepath.Join(songDir, songResp.Path),
		IsPlayed:       false,
	}
	return song, songResp.Id, true
}

func downloadSongFile(url, filePath, username, password string) error {
	req, _ := http.NewRequest("GET", url, nil)
	setAuthHeader(req, username, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		log.Fatalln("Song not found")
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Return code is not 200: %d\n", resp.StatusCode)
	}

	fh, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalln(err)
		}
		if n > 0 {
			_, err = fh.Write(buf[:n])
			if err != nil {
				log.Fatalln(err)
			}
		}
		if err == io.EOF || n == 0 {
			break
		}
	}

	resp.Body.Close()
	fh.Close()

	return nil
}
