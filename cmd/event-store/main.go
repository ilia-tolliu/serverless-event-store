package main

import (
	"github.com/ilia-tolliu-go-event-store/internal/web"
	"log"
	"net/http"
)

func main() {
	log.Println("Hello!")

	webApp := web.NewEsWebApp()

	server := http.Server{
		Addr:    ":8080",
		Handler: webApp,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen and serve: %s", err)
	}
}
