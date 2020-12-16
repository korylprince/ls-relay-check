package main

import (
	"log"
	"net/http"
)

func main() {
	c := NewConfig()
	log.Println("Listening on:", c.ListenAddr)
	http.ListenAndServe(c.ListenAddr, http.StripPrefix(c.Prefix, c.Router()))
}
