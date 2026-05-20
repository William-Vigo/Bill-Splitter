package routes

import (
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/response"
	"github.com/William-Vigo/Bill-Splitter/worker"
	"github.com/gorilla/mux"
)

func EndPoints(r *mux.Router) {
	r.Handle("/split-bill", response.Handler(worker.WorkerHandler)).Methods(http.MethodPost)
}
