package main

import (
	// "fmt"
	"log"
	"net/http"

	"github.com/onbehalfofhim/metric-alert/internal/handler"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

func main() {
	storage := models.NewMemStorage()

	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.RootHandler(*storage, w, r)
		},
	)))
}
