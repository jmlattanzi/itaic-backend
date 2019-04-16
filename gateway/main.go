package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"goji.io"
	"goji.io/pat"
)

func main() {
	fmt.Println("[ * ] Gateway starting....")

	router := goji.NewMux()

	router.HandleFunc(pat.Get("/api/posts"), func(res http.ResponseWriter, req *http.Request) {
		result, err := http.Get("http://176.24.0.13/posts")
		if err != nil {
			fmt.Println("[ ! ] Error calling endpoint")
		}
		defer result.Body.Close()

		json.NewEncoder(res).Encode(&result.Body)
	})

	http.ListenAndServe(":6000", handlers.LoggingHandler(os.Stdout, router))
}
