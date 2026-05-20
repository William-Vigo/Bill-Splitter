package response

import "net/http"

type Handler func(w http.ResponseWriter, r *http.Request) error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		//TODO write back a response
	}
}
