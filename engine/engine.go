package main

import (
	"github.com/jhcook/game_engine/hangman"
	"log"
	"net/http"
)

func apiHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: %v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func apiHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return apiHandler(next)
}

func main() {
	// Create the singleton "men" in hangman to keep track of all hangman episodes
	hangman.NewMen()
	//    http.Handle("/list", apiHandlerFunc())
	//    http.Handle("/new", apiHandlerFunc())
	//    http.Handle("/status", apiHandlerFunc())
	http.Handle("/hangman", apiHandlerFunc(hangman.Play))
	http.ListenAndServe(":3000", nil)
}
