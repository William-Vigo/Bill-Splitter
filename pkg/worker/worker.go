package worker

import (
	"fmt"
	"net/http"
)

func WorkerHandler(w http.ResponseWriter, r *http.Request) error {

	fmt.Fprintf(w, "works")
	return nil
}
