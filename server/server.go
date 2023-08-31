package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	log.Printf("Received %s request on %s\n", req.Method, req.URL)

	idOrUrl := req.URL.Query().Get("idorurl")
	if idOrUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(metadata)
	if err != nil {
		log.Fatal("Error parsing JSON.\n")
	}
	w.Write(jsonResp)
}
