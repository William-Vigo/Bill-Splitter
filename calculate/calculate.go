package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type Payload struct {
	People  []People      `json:"people"`
	Shared  []SharedItems `json:"sharedItems"`
	TipPaid float64       `json:"tipPaid"`
	TaxPaid float64       `json:"taxPaid"`
}

type People struct {
	Name      string `json:"name"`
	Purchases []Item `json:"items"`
}

type Item struct {
	Name  string  `json:"itemName"`
	Price float64 `json:"price"`
}

type SharedItems struct {
	People []struct {
		Name string `json:"name"`
	} `json:"people"`
	Purchases []Item `json:"items"`
}

type Receipt struct {
	Name        string          `json:"name"`
	Items       []Item          `json: "items"`
	SharedItems []Item          `json: "sharedItems"`
	ItemSum     decimal.Decimal `json:"itemSum"`
	Tax         decimal.Decimal `json:"tax"`
	Tip         decimal.Decimal `json:"tip"`
	Total       decimal.Decimal `json:"total"`
}

type Result struct {
	Receipt   map[string]Receipt
	BillTotal decimal.Decimal
	Data      Payload

	// Required so that items are calculated accordingly
	taxPercentage decimal.Decimal
	tipPercentage decimal.Decimal
}

func (r *Result) processIndividual() {
	for _, person := range r.Data.People {
		for _, item := range person.Purchases {
			receipt := r.Receipt[person.Name]
			receipt.ItemSum = receipt.ItemSum.Add(decimal.NewFromFloat(item.Price))
			receipt.Items = append(receipt.Items, item)
			r.Receipt[person.Name] = receipt
		}
	}

}
func (r *Result) processSplit() {
	if len(r.Data.Shared) == 0 {
		return
	}

	for _, group := range r.Data.Shared {
		for _, item := range group.Purchases {
			val := decimal.NewFromFloat(item.Price).Div(decimal.NewFromInt(int64(len(group.People))))
			for _, people := range group.People {
				receipt := r.Receipt[people.Name]
				receipt.SharedItems = append(receipt.SharedItems, Item{
					Name:  item.Name,
					Price: val.InexactFloat64(),
				})
				receipt.ItemSum = receipt.ItemSum.Add(val)
				r.Receipt[people.Name] = receipt
			}
		}

	}
}
func (r *Result) calculatePercentages() {
	total := decimal.NewFromFloat(0)
	// calculate itemTotal
	for _, val := range r.Receipt {
		total = total.Add(val.ItemSum)
	}
	r.taxPercentage = decimal.NewFromFloat(r.Data.TaxPaid).Div(total)
	r.tipPercentage = decimal.NewFromFloat(r.Data.TipPaid).Div(total)
}

func (r *Result) postProcess() {
	for key, val := range r.Receipt {
		tax := val.ItemSum.Mul(r.taxPercentage)
		tip := val.ItemSum.Mul(r.tipPercentage)

		receipt := r.Receipt[key]
		receipt.Name = key
		receipt.Tax = tax
		receipt.Tip = tip
		receipt.Total = tax.Add(tip).Add(receipt.ItemSum)
		receipt.ItemSum = val.ItemSum

		r.BillTotal = r.BillTotal.Add(receipt.Total)
		r.Receipt[key] = receipt

	}
}

func (r *Result) Process() string {

	r.Receipt = make(map[string]Receipt)
	r.BillTotal = decimal.NewFromFloat(0.0)

	// pre process
	r.processIndividual()
	r.processSplit()

	// post process
	r.calculatePercentages()
	r.postProcess()

	output, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		fmt.Printf("Error marshalling map: %v\n", err.Error())
	}

	return string(output)
}

func (mapping *Result) MarshalJSON() ([]byte, error) {
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
