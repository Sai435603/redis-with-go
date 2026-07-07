package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Hello world!!")
	router := chi.NewRouter()

	http.ListenAndServe(":8000", router)
}
