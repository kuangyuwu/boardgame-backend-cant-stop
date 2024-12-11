package main

import (
	"flag"
	"fmt"
	"log"

	server "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/server"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	l := server.InitializeLobby()
	srv := server.InitializeServer(addr, l)
	fmt.Println("Starting server on address", *addr)
	log.Fatal(srv.ListenAndServe())
}
