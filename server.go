package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func initializeServer(addr *string, l *Lobby) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthz", handlerReadiness)
	mux.HandleFunc("/", l.handlerDefault)

	return &http.Server{
		Addr:    *addr,
		Handler: mux,
	}
}

func (l *Lobby) handlerDefault(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "http://cant-stop.kuangyuwu.com"
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade request failed: %s\n", err)
		return
	}

	u, err := l.createUser(conn)
	if err != nil {
		log.Printf("Error creating user: %s\n", err)
		conn.Close()
		return
	}

	go u.handleMessage()
	go u.sendMessage()
	log.Print("a user connected")
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
