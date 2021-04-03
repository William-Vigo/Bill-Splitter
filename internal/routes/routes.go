package routes

import (
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/internal/response"
	"github.com/William-Vigo/Bill-Splitter/pkg/worker"
	"github.com/gorilla/mux"
)

func EndPoints(r *mux.Router) {
	r.Handle("/split-bill", response.Handler(worker.WorkerHandler)).Methods(http.MethodGet)
}
