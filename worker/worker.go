package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/calculate"
)

func WorkerHandler(w http.ResponseWriter, r *http.Request) error {
	body, _ := ioutil.ReadAll(r.Body)
	var payload calculate.Payload
	json.Unmarshal(body, &payload)

	//TODO return err
	result := calculate.Result{
		Data: payload,
	}
	output := result.Process()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, output)
	fmt.Println("Processed bill split request")
	return nil
}
