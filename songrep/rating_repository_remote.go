package songrep

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type RemoteRatingRepository struct {
	ServerBaseUrl string
	Username      string
	Password      string
}

func (r *RemoteRatingRepository) AddPlay(song Song, timestamp int, rating int) {
	songId := computeSongId(song)
	url := r.ServerBaseUrl + "/api/songs/" + songId + "/play/"
	body := fmt.Sprintf("{\"timestamp\":%d,\"rating\":%d}", timestamp, rating)
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.Username, r.Password)))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Status code is not 200: %d\n", resp.StatusCode)
	}
}

func computeSongId(song Song) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(song.Path)))
}
