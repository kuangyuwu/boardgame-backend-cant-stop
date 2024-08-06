package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	respondWithJSON(w, http.StatusOK, payload)
}

func handlerConnect(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Implement your origin check logic here
			// For example, allow only specific origins
			origin := r.Header.Get("Origin")
			return origin == "http://127.0.0.1:5500"
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade request failed: %s\n", err)
		return
	}
	defer conn.Close()

	gameIdString := r.PathValue("gameId")
	fmt.Printf("%s\n", gameIdString)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %s\n", err)
			}
			break
		}
		if len(msg) > 0 {
			log.Print(string(msg))
			break
		}
	}
	payload := struct {
		Test string `json:"test"`
	}{
		Test: "test456",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON %s", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
	log.Print("Done")
}

func main() {
	flag.Parse()
	fmt.Println("Starting server on address", *addr)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthz", handlerReadiness)
	mux.HandleFunc("/v1/connect/{gameId}", handlerConnect)

	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))

}
