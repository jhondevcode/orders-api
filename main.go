package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Hello, World!")

	server := &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(basicHandler),
	}
	log.Fatalln(server.ListenAndServe())
}

func basicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
