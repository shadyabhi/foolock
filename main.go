package main

import (
	"log"
	"net/http"

	"github.com/shadyabhi/foolock/lockstate"
	"github.com/shadyabhi/foolock/lockstatehttp"
)

const ServerAddr = ":8080"

func main() {
	state := lockstate.New()
	handler := lockstatehttp.NewHandler(state)

	http.HandleFunc("/lock", handler.HandleLock)

	log.Printf("Starting lock service on %s", ServerAddr)
	if err := http.ListenAndServe(ServerAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
