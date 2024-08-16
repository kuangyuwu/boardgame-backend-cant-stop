package main

import (
	"flag"
	"fmt"
	"log"
)

var addr = flag.String("addr", ":80", "http service address")

func main() {
	flag.Parse()

	l := initializeLobby()
	srv := initializeServer(addr, l)
	fmt.Println("Starting server on address", *addr)
	log.Fatal(srv.ListenAndServe())
}
