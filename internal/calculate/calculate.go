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
	Total   float64 `json:"total"`
}

type CustomMap struct {
	Receipt   map[string]Receipt
	BillTotal float64
}

func Process(data Payload) string {
	moneyOwed := CustomMap{
		Receipt:   make(map[string]Receipt),
		BillTotal: 0.00,
	}

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
					receipt := moneyOwed.Receipt[person.Name]
					receipt.ItemSum = utility.Round(receipt.ItemSum+purchases.Price, 2)
					moneyOwed.Receipt[person.Name] = receipt
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

					moneyDue := utility.Round(total/(float64(splitSize)), 2)

					for _, people := range group.People {
						lock.Lock()
						receipt := moneyOwed.Receipt[people.Name]
						receipt.ItemSum = utility.Round(receipt.ItemSum+moneyDue, 2)
						moneyOwed.Receipt[people.Name] = receipt
						lock.Unlock()
					}

				}(group)

			}
		}()
	}

	wg.Wait()

	for key, val := range moneyOwed.Receipt {

		// Tax calculate
		tax := utility.Round(val.ItemSum*constants.TaxRate, 2)
		tip := utility.Round(val.ItemSum*data.TipPercent, 2)
		receipt := moneyOwed.Receipt[key]
		receipt.Name = key
		receipt.Tax = tax
		receipt.Tip = tip
		receipt.Total = utility.Round(tax+tip+receipt.ItemSum, 2)
		moneyOwed.BillTotal = utility.Round(moneyOwed.BillTotal+receipt.Total, 2)
		moneyOwed.Receipt[key] = receipt

	}

	output, err := json.MarshalIndent(moneyOwed, "", "\t")
	if err != nil {
		fmt.Printf("Error marshalling map: %v\n", err.Error())
	}

	return string(output)
}

func (mapping CustomMap) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{\"people\":[")
	length := len(mapping.Receipt)
	count := 0
	for _, val := range mapping.Receipt {
		jsonVal, _ := json.Marshal(val)

		buffer.WriteString(string(jsonVal))
		count++
		if count < length {

			buffer.WriteString(",")
		}
	}
	buffer.WriteString("],")
	buffer.WriteString(fmt.Sprintf("\"billTotal\": %.2f}", mapping.BillTotal))
	return buffer.Bytes(), nil
}
