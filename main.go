package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	fmt.Println("Hello world!!")
	router := chi.NewRouter()

	http.ListenAndServe(":8000", router)
}
