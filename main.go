package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"yt_search_server/server"
	"yt_search_server/youtube"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	youtubeApiKey := os.Getenv("YOUTUBE_DATA_SERVICE_API_KEY")

	youtube := youtube.NewYouTubeService(youtubeApiKey)
	server := server.NewServer(youtube)

	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/", server.GetHome).Methods("GET")
	router.HandleFunc("/metadata/", server.GetMetadata).Methods("GET")

	serveUrl := os.Getenv("SERVER_HOST") + ":" + os.Getenv("SERVER_PORT")

	log.Printf("Server listening for connections at %s\n", serveUrl)

	err = http.ListenAndServe(serveUrl, router)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed.")
	} else {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
