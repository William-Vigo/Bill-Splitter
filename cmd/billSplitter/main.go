package main

import (
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/internal/routes"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)

	routes.EndPoints(router)
	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:3000",
	}

	srv.ListenAndServe()
}
