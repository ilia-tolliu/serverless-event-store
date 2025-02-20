package main

import (
	"github.com/ilia-tolliu-go-event-store/internal/web"
	"log"
	"net/http"
)

func main() {
	log.Println("Hello!")

	router := web.NewRouter()

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen and serve: %s", err)
	}
}
