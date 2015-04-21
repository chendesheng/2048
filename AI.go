package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	http.HandleFunc("/", AI)
	log.Fatal(http.ListenAndServe("localhost:8877", nil))
}

func AI(w http.ResponseWriter, req *http.Request) {
	nextmove := rand.Int() % 4
	fmt.Printf("next move:%d\n", nextmove)
	io.WriteString(w, fmt.Sprintf("move(%d);", nextmove))
}
