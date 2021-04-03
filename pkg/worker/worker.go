package worker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/William-Vigo/Bill-Splitter/internal/calculate"
)

func WorkerHandler(w http.ResponseWriter, r *http.Request) error {
	body, _ := ioutil.ReadAll(r.Body)
	payload := calculate.Payload{}
	json.Unmarshal(body, &payload)

	calculate.Process(payload)

	return nil
}
