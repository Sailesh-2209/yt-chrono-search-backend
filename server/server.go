package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"yt_search_server/youtube"
)

type Server struct {
	youtube *youtube.YouTube
}

func NewServer(youtube *youtube.YouTube) *Server {
	return &Server{youtube: youtube}
}

func (server *Server) GetHome(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received %s request on %s\n", req.Method, req.URL)
	fmt.Fprint(w, "Welcome to the YouTube Search Server!")
}

func (server *Server) GetMetadata(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Received %s request on %s\n", req.Method, req.URL)

	idOrUrl := req.URL.Query().Get("idorurl")
	if idOrUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := make(map[string]string)
		resp["message"] = "Bad Request. Query parameter 'idorurl' missing."
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error parsing JSON.\n")
		}
		w.Write(jsonResp)
		return
	}

	metadata, err := server.youtube.GetVideoMetadata(idOrUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprintf("Error while fetching video metadata: %s", err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error parsing JSON.\n")
		}
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, err := json.Marshal(metadata)
	if err != nil {
		log.Fatal("Error parsing JSON.\n")
	}
	w.Write(jsonResp)
}

func (server *Server) GetVideos(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Received %s request on %s\n", req.Method, req.URL)

	qpChannelId := req.URL.Query().Get("channelId")
	if qpChannelId == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := make(map[string]string)
		resp["message"] = "Bad Request. Query parameter 'id' missing."
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error forming JSON.\n")
		}
		w.Write(jsonResp)
		return
	}

	qpVideoId := req.URL.Query().Get("videoId")
	if qpVideoId == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := make(map[string]string)
		resp["message"] = "Bad Request. Query parameter 'videoId' missing."
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error forming JSON.\n")
		}
		w.Write(jsonResp)
		return
	}

	qpPageToken := req.URL.Query().Get("pageToken")

	loadPrev := false
	qpLoadPrev := req.URL.Query().Get("loadPrev")
	if qpLoadPrev == "" || strings.ToLower(qpLoadPrev) == "false" {
		loadPrev = false
	} else if strings.ToLower(qpLoadPrev) == "true" {
		loadPrev = true
	} else {
		w.WriteHeader(http.StatusBadRequest)
		resp := make(map[string]string)
		resp["message"] = "Bad Request. Invalid value for query parameter 'videoId'. If defined, it must be one of 'true' or 'false'."
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error forming JSON.\n")
		}
		w.Write(jsonResp)
		return
	}

	videos, err := server.youtube.GetChannelVideos(
		qpChannelId, qpVideoId, qpPageToken, loadPrev,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprintf("Error while fetching videos: %s", err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatal("Error forming JSON.\n")
		}
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonResp, err := json.Marshal(videos)
	if err != nil {
		log.Fatal("Error forming JSON.\n")
	}
	w.Write(jsonResp)
}
