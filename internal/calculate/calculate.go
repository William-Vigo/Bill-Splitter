package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/William-Vigo/Bill-Splitter/internal/constants"
	"github.com/William-Vigo/Bill-Splitter/pkg/worker/utility"
)

type Payload struct {
	Group      []People      `json:"people"`
	Shared     []SharedItems `json:"sharedItems"`
	TipPercent float64       `json:"tipPercentage"`
}

type People struct {
	Name      string  `json:"name"`
	Purchases []Items `json:"items"`
}

type Items struct {
	Name  string  `json:"itemName"`
	Price float64 `json:"price"`
}

type SharedItems struct {
	People []struct {
		Name string `json:"name"`
	} `json:"people"`
	Purchases []Items `json:"items"`
}

type Receipt struct {
	Name    string  `json:"name"`
	ItemSum float64 `json:"itemSum"`
	Tax     float64 `json:"tax"`
	Tip     float64 `json:"tip"`
	Total   float64 `json:"total`
}

type CustomMap map[string]Receipt

func Process(data Payload) string {
	moneyOwed := make(CustomMap)

	var wg sync.WaitGroup
	var lock sync.RWMutex
	// this sums individuals total
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, person := range data.Group {
			wg.Add(1)
			go func(person People) {
				defer wg.Done()
				for _, purchases := range person.Purchases {
					lock.Lock()
					receipt := moneyOwed[person.Name]
					receipt.ItemSum += purchases.Price
					moneyOwed[person.Name] = receipt
					lock.Unlock()
				}
			}(person)
		}
	}()

	if len(data.Shared) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, group := range data.Shared {
				wg.Add(1)
				go func(group SharedItems) {
					defer wg.Done()

					splitSize := len(group.People)
					total := 0.00
					for _, purchases := range group.Purchases {
						total += purchases.Price
					}

					moneyDue := utility.Round(total/float64(splitSize), 2)
					//TODO round money to 2 decimal places

					for _, people := range group.People {
						lock.Lock()
						receipt := moneyOwed[people.Name]
						receipt.ItemSum += moneyDue
						moneyOwed[people.Name] = receipt
						lock.Unlock()
					}

				}(group)

			}
		}()
	}

	wg.Wait()

	for key, val := range moneyOwed {

		// Tax calculate
		tax := utility.Round(val.ItemSum*constants.TaxRate, 2)
		tip := utility.Round(val.ItemSum*data.TipPercent, 2)

		receipt := moneyOwed[key]
		receipt.Name = key
		receipt.Tax = tax
		receipt.Tip = tip
		receipt.Total = utility.Round(tax+tip+receipt.ItemSum, 2)
		moneyOwed[key] = receipt

	}

	output, err := json.Marshal(moneyOwed)
	if err != nil {
		fmt.Printf("Error marshalling map: %v\n", err.Error())
	}

	return string(output)
}

func (mapping CustomMap) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"people\":[")
	length := len(mapping)
	count := 0
	for _, val := range mapping {
		jsonVal, _ := json.Marshal(val)

		buffer.WriteString(string(jsonVal))
		count++
		if count < length {

			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]}")
	return buffer.Bytes(), nil
}
