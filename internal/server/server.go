package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
)

func New(addr *string, m WebsocketManager) (*http.Server, <-chan bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerHealthz)
	mux.HandleFunc("/healthz", handlerHealthz)
	handlerWebsocket := generateHandlerWebsocket(m)
	mux.HandleFunc("/websocket", handlerWebsocket)

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	done := make(chan bool)
	go gracefulShutdown(srv, done)

	return srv, done
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
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
		clog.Errorf("error marshalling JSON %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(data); err != nil {
		clog.Errorf("failed to write response: %v", err)
	}
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	clog.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		clog.Errorf("server forced to shutdown with error: %v", err)
	}

	clog.Info("server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}
