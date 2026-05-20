package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
)

type Payload struct {
	Group   []People      `json:"people"`
	Shared  []SharedItems `json:"sharedItems"`
	TipPaid float64       `json:"tipPaid"`
	TaxPaid float64       `json:"taxPaid"`
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
	Name    string          `json:"name"`
	ItemSum decimal.Decimal `json:"itemSum"`
	Tax     decimal.Decimal `json:"tax"`
	Tip     decimal.Decimal `json:"tip"`
	Total   decimal.Decimal `json:"total"`
}

type CustomMap struct {
	Receipt   map[string]Receipt
	BillTotal decimal.Decimal
}

func Process(data Payload) string {
	moneyOwed := CustomMap{
		Receipt:   make(map[string]Receipt),
		BillTotal: decimal.NewFromFloat(0.0),
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
					receipt.ItemSum = receipt.ItemSum.Add(decimal.NewFromFloat(purchases.Price))
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
					total := decimal.NewFromInt(0)
					for _, purchases := range group.Purchases {
						total = total.Add(decimal.NewFromFloat(purchases.Price))
					}
					moneyDue := total.Div(decimal.NewFromInt(int64(splitSize)))

					for _, people := range group.People {
						lock.Lock()
						receipt := moneyOwed.Receipt[people.Name]
						receipt.ItemSum = receipt.ItemSum.Add(moneyDue)
						moneyOwed.Receipt[people.Name] = receipt
						lock.Unlock()
					}

				}(group)

			}
		}()
	}

	wg.Wait()
	total := decimal.NewFromFloat(0)
	// calculate itemTotal
	for _, val := range moneyOwed.Receipt {
		total = total.Add(val.ItemSum)
	}
	calculatedTaxPercentage := decimal.NewFromFloat(data.TaxPaid).Div(total)
	calculatedTipPercentage := decimal.NewFromFloat(data.TipPaid).Div(total)
	for key, val := range moneyOwed.Receipt {
		// Tax calculate
		tax := val.ItemSum.Mul(calculatedTaxPercentage)
		tip := val.ItemSum.Mul(calculatedTipPercentage)
		receipt := moneyOwed.Receipt[key]
		receipt.Name = key
		receipt.Tax = tax
		receipt.Tip = tip
		receipt.Total = tax.Add(tip).Add(receipt.ItemSum)
		receipt.ItemSum = val.ItemSum
		moneyOwed.BillTotal = moneyOwed.BillTotal.Add(receipt.Total)
		moneyOwed.Receipt[key] = receipt

	}

	output, err := json.MarshalIndent(moneyOwed, "", "\t")
	if err != nil {
		fmt.Printf("Error marshalling map: %v\n", err.Error())
	}

	return string(output)
}

func (mapping CustomMap) MarshalJSON() ([]byte, error) {
	decimal.MarshalJSONWithoutQuotes = true
	buffer := bytes.NewBufferString("{\"people\":[")
	length := len(mapping.Receipt)
	count := 0
	itemTotal := decimal.NewFromInt(0)
	taxTotal := decimal.NewFromInt(0)
	tipTotal := decimal.NewFromInt(0)
	for _, val := range mapping.Receipt {
		taxTotal = taxTotal.Add(val.Tax)
		tipTotal = tipTotal.Add(val.Tip)
		itemTotal = itemTotal.Add(val.ItemSum)
	}
	for _, val := range mapping.Receipt {
		val.ItemSum = val.ItemSum.Round(2)
		val.Tax = val.Tax.Round(2)
		val.Tip = val.Tip.Round(2)
		val.Total = val.Total.Round(2)

		jsonVal, _ := json.Marshal(val)
		buffer.WriteString(string(jsonVal))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	FinalItemTotal, _ := itemTotal.Round(2).Float64()
	FinalTaxTotal, _ := taxTotal.Round(2).Float64()
	FinalTipTotal, _ := tipTotal.Round(2).Float64()
	FinalBillTotal, _ := mapping.BillTotal.Round(2).Float64()
	buffer.WriteString("],")
	buffer.WriteString(fmt.Sprintf("\"itemTotal\": %.2f,\n", FinalItemTotal))
	buffer.WriteString(fmt.Sprintf("\"taxTotal\": %.2f,\n", FinalTaxTotal))
	buffer.WriteString(fmt.Sprintf("\"tipTotal\": %.2f,\n", FinalTipTotal))
	buffer.WriteString(fmt.Sprintf("\"billTotal\": %.2f}", FinalBillTotal))
	return buffer.Bytes(), nil
}
