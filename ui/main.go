package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gptscript-ai/otto/apiclient"
	"github.com/gptscript-ai/otto/ui/router"
)

func main() {
	client := &apiclient.Client{
		BaseURL: "http://localhost:8080",
		Token:   os.Getenv("OTTO_TOKEN"),
	}
	handler := router.Init(client, true)
	log.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", handler)
	if err != nil {
		log.Fatal(err)
	}
}
