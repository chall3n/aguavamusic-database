package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Spotify API
const (
	spotifyAPI = "https://api.spotify.com/v1"
	tokenURL   = "https://accounts.spotify.com/api/token"
)

// Song data structure
type song struct {
	Name       string `json:"name"`
	Streams    uint   `json:"streams"`
	Key        string `json:"key"`
	Bpm        uint   `json:"bpm"`
	SpotifyID  string `json:"spotifyId"`
	Popularity int    `json:"popularity"` // This is the only dynamically updated parameter
}

var songs = []song{
	{Name: "If I", Streams: 924000, Key: "B", Bpm: 119, SpotifyID: "54Ew6UcuXLChTnSAwXAIXY"},
	{Name: "Payday", Streams: 556000, Key: "F", Bpm: 126, SpotifyID: "4gpOjiawQcmFqRSwtp7Ppt"},
	{Name: "Quédate Conmigo", Streams: 111000, Key: "C", Bpm: 120, SpotifyID: "23Byo25q7SjXvpumZv3q4K"},
	{Name: "Whatever Happens, Happens", Streams: 472000, Key: "F#", Bpm: 118, SpotifyID: "2APRTIVViZequi0ZClGao5"},
	{Name: "Silk Thieves", Streams: 85900, Key: "G", Bpm: 125, SpotifyID: "3NX8oez6NU2BjJHDK06a65"},
	{Name: "Sinnerman", Streams: 87100, Key: "A#", Bpm: 127, SpotifyID: "2WSlfpTj1WD3H6vKZCXyik"},
	{Name: "Vanilla", Streams: 140000, Key: "F#", Bpm: 116, SpotifyID: "0KNQTHbKpmQtRSDgYhkJf7"},
	{Name: "Can We", Streams: 94800, Key: "F#", Bpm: 122, SpotifyID: "7lDSJLLBjsfPhZi79APe6i"},
	{Name: "Killin' Me", Streams: 13501, Key: "A#", Bpm: 126, SpotifyID: "5u1SCoT3vRGHnpwYATaNUD"},
	{Name: "Aquatic Movements", Streams: 19500, Key: "A", Bpm: 124, SpotifyID: "1nHYUvlvvT57PtXdunIJP1"},
}

// Spotify Token Caching
var (
	cachedToken     string
	tokenExpiration time.Time
	tokenMutex      sync.Mutex
)

// Fetches a token using Client Credentials Flow, caches it so we don't have to fetch everytime
func getSpotifyAccessToken(clientID, clientSecret string) (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if cachedToken != "" && time.Now().Before(tokenExpiration.Add(-60*time.Second)) {
		return cachedToken, nil
	}

	authString := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encodedAuthString := base64.StdEncoding.EncodeToString([]byte(authString))

	data := []byte("grant_type=client_credentials")

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Add("Authorization", "Basic "+encodedAuthString)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch token: status %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	if response.AccessToken == "" {
		return "", fmt.Errorf("received empty access token")
	}

	cachedToken = response.AccessToken
	tokenExpiration = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)

	return cachedToken, nil
}

// Makes request to Spotify and includes basic rate limit handling
func doSpotifyRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	maxRetries := 3
	retryCount := 0

	for retryCount < maxRetries {
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed executing request to Spotify: %w", err)
		}

		// Check for rate limit (HTTP 429)
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			retryAfter := resp.Header.Get("Retry-After")
			retrySeconds, parseErr := strconv.Atoi(retryAfter)
			if parseErr != nil {
				retrySeconds = 5 + retryCount*2
			}
			if retryCount >= maxRetries-1 {
				return nil, fmt.Errorf("rate limit exceeded after %d retries", maxRetries)
			}
			fmt.Printf("Rate limit exceeded. Retrying after %d seconds...\n", retrySeconds)
			time.Sleep(time.Duration(retrySeconds) * time.Second)
			retryCount++
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("max retries reached for Spotify request")
}

// Fetches details for a batch of tracks on Spotify by ID
func getTracksDetails(trackIDs []string, accessToken string) (map[string]int, error) {
	if len(trackIDs) == 0 {
		return make(map[string]int), nil
	}
	if len(trackIDs) > 50 {
		return nil, fmt.Errorf("too many track IDs provided to getTracksDetails; max is 50, got %d", len(trackIDs))
	}

	url := spotifyAPI + "/tracks?ids=" + strings.Join(trackIDs, ",")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracks details request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := doSpotifyRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracks details response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch track data: status %d - %s", resp.StatusCode, string(body))
	}

	var data struct {
		Tracks []struct {
			ID         string `json:"id"`
			Popularity int    `json:"popularity"`
		} `json:"tracks"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal track data (response body: %s): %w", string(body), err)
	}

	popularityMap := make(map[string]int)
	for _, track := range data.Tracks {
		if track.ID != "" {
			popularityMap[track.ID] = track.Popularity
		}
	}

	return popularityMap, nil
}

// Gin Handler
// getSongs fetches song data, gets current popularity from Spotify, sorts, and returns
func getSongs(c *gin.Context) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	sortBy := c.DefaultQuery("sortBy", "streams")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	accessToken, err := getSpotifyAccessToken(clientID, clientSecret)
	if err != nil {
		fmt.Println("Error getting Spotify access token:", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Could not authenticate with Spotify"})
		return
	}

	var idsToFetch []string
	for _, s := range songs {
		if s.SpotifyID != "" && !strings.HasPrefix(s.SpotifyID, "YOUR_REAL_") {
			idsToFetch = append(idsToFetch, s.SpotifyID)
		}
	}

	allPopularities := make(map[string]int)
	batchSize := 50

	fmt.Printf("Attempting to fetch popularity for %d track IDs...\n", len(idsToFetch))

	for i := 0; i < len(idsToFetch); i += batchSize {
		end := i + batchSize
		if end > len(idsToFetch) {
			end = len(idsToFetch)
		}
		batchIDs := idsToFetch[i:end]

		fmt.Printf("Fetching batch %d-%d: %v\n", i, end-1, batchIDs)
		popularityMap, err := getTracksDetails(batchIDs, accessToken)
		if err != nil {
			fmt.Printf("Error fetching details for batch %v: %v\n", batchIDs, err)
		} else {
			for id, pop := range popularityMap {
				allPopularities[id] = pop
			}
		}
	}
	fmt.Printf("Finished fetching popularities. Result map size: %d\n", len(allPopularities))

	responseSongs := make([]song, len(songs))
	copy(responseSongs, songs)

	foundCount := 0
	for i := range responseSongs {
		if responseSongs[i].SpotifyID != "" {
			popularity, found := allPopularities[responseSongs[i].SpotifyID]
			if found {
				responseSongs[i].Popularity = popularity
				foundCount++
			} else {
				responseSongs[i].Popularity = 0
			}
		} else {
			responseSongs[i].Popularity = 0
		}
	}
	fmt.Printf("Successfully mapped popularity for %d tracks.\n", foundCount)

	switch sortBy {
	case "streams":
		if sortOrder == "asc" {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Streams < responseSongs[j].Streams })
		} else {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Streams > responseSongs[j].Streams })
		}
	case "key":
		if sortOrder == "asc" {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Key < responseSongs[j].Key })
		} else {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Key > responseSongs[j].Key })
		}
	case "bpm":
		if sortOrder == "asc" {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Bpm < responseSongs[j].Bpm })
		} else {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Bpm > responseSongs[j].Bpm })
		}
	case "popularity":
		if sortOrder == "asc" {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Popularity < responseSongs[j].Popularity })
		} else {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Popularity > responseSongs[j].Popularity })
		}
	case "name":
		if sortOrder == "asc" {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Name < responseSongs[j].Name })
		} else {
			sort.Slice(responseSongs, func(i, j int) bool { return responseSongs[i].Name > responseSongs[j].Name })
		}
	}

	c.IndentedJSON(http.StatusOK, responseSongs)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Frontend URL
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/songs", getSongs)

	fmt.Println("Go API server running on localhost:8080")
	err = router.Run("localhost:8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
