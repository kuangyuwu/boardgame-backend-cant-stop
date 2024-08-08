package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func (cfg *Config) initializeServer(addr *string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthz", handlerReadiness)
	mux.HandleFunc("/v1/user/{username}", cfg.handlerUser)

	cfg.server = &http.Server{
		Addr:    *addr,
		Handler: mux,
	}
}

func (cfg *Config) handlerUser(w http.ResponseWriter, r *http.Request) {

	username := r.PathValue("username")
	if len(username) > MaxLenUsername {
		log.Print("The username is too long")
		respondWithError(w, http.StatusBadRequest, "The username is too long")
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "http://127.0.0.1:5500"
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade request failed: %s\n", err)
		return
	}

	user, err := cfg.createUser(username, conn)
	if err != nil {
		log.Printf("Error creating user: %s\n", err)
		return
	}

	go cfg.handle(user)
	go user.sendMessage()

	log.Print("User created")
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	respondWithJSON(w, http.StatusOK, payload)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	payload := errorResponse{
		Error: msg,
	}
	respondWithJSON(w, code, payload)
}
