package main

import (
	"fmt"
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/routes"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)

	routes.EndPoints(router)
	srv := &http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	fmt.Println("Starting up bill-splitter server")
	srv.ListenAndServe()
}
