package main

import (
	"flag"
	"net/http"

	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/clog"
	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/lobby"
	"github.com/kuangyuwu/boardgame-backend-cant-stop/internal/server"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	l := lobby.New()
	srv, done := server.New(addr, l)
	clog.Infof("starting server on address %s", *addr)

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		clog.Errorf("http server error: %s", err)
		panic("")
	}

	// Wait for the graceful shutdown to complete
	<-done
	clog.Info("graceful shutdown complete")
}
