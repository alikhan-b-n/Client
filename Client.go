package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	_ "log"
	"net/http"
)

type Video struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Client struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

var videos []Video
var clients []Client

// Middleware to check if client is logged in
func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if the token is valid
		var client *Client
		for _, c := range clients {
			if c.Token == token {
				client = &c
				break
			}
		}

		if client == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Pass the authenticated client to the next handler
		r = r.WithContext(context.WithValue(r.Context(), "client", client))
		next(w, r)
	}
}

// Handler for creating a new video
func createVideo(w http.ResponseWriter, r *http.Request) {
	var newVideo Video
	err := json.NewDecoder(r.Body).Decode(&newVideo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Generate a unique ID for the video
	newVideo.ID = fmt.Sprintf("video%d", len(videos)+1)

	// Add the video to the list
	videos = append(videos, newVideo)

	// Return the created video as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newVideo)
}

// Handler for retrieving all videos
func getVideos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

// Handler for retrieving a specific video by ID
func getVideo(w http.ResponseWriter, r *http.Request) {
	videoID := r.URL.Query().Get("id")
	for _, video := range videos {
		if video.ID == videoID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(video)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// Handler for updating a video by ID
func updateVideo(w http.ResponseWriter, r *http.Request) {
	videoID := r.URL.Query().Get("id")
	var updatedVideo Video
	err := json.NewDecoder(r.Body).Decode(&updatedVideo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i, video := range videos {
		if video.ID == videoID {
			// Update the video's details
			video.Title = updatedVideo.Title
			video.URL = updatedVideo.URL

			// Update the video in the list
			videos[i] = video

			// Return the updated video as JSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(video)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func deleteVideo(w http.ResponseWriter, r *http.Request) {
	videoID := r.URL.Query().Get("id")
	for i, video := range videos {
		if video.ID == videoID {
			// Remove the video from the list
			videos = append(videos[:i], videos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	var client Client
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate client credentials
	for i, c := range clients {
		if c.Username == client.Username && c.Password == client.Password {
			// Generate an authentication token
			token := GenerateSecureToken(len(c.Username))

			// Update the client's token
			clients[i].Token = token

			// Return the token as JSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"token": token})
			return
		}
	}

	w.WriteHeader(http.StatusUnauthorized)
}

// Handler for user registration
func register(w http.ResponseWriter, r *http.Request) {
	var client Client
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the username is already taken
	for _, c := range clients {
		if c.Username == client.Username {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	// Generate a unique ID for the client
	client.ID = fmt.Sprintf("client%d", len(clients)+1)

	// Add the client to the list
	clients = append(clients, client)

	w.WriteHeader(http.StatusCreated)
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func main() {
	// Set up the HTTP server and routes
	http.HandleFunc("/videos", authenticate(getVideos))
	http.HandleFunc("/video", authenticate(getVideo))
	http.HandleFunc("/video/create", authenticate(createVideo))
	http.HandleFunc("/video/update", authenticate(updateVideo))
	http.HandleFunc("/video/delete", authenticate(deleteVideo))
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
