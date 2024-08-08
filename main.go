package main

import (
	"flag"
	"fmt"
	"log"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	config := initializeConfig()

	config.initializeServer(addr)
	fmt.Println("Starting server on address", *addr)
	log.Fatal(config.server.ListenAndServeTLS("server.crt", "server.key"))

}
