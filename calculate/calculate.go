package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
)

type currency int64

func (c currency) Value() *money.Money {
	return money.New(int64(c), money.USD)
}

func (c currency) Add(val currency) currency {
	sum, err := c.Value().Add(val.Value())
	if err != nil {
		panic(err.Error())
	}
	return currency(sum.Amount())
}

func (c currency) Split(n int) []currency {
	split, err := c.Value().Split(n)
	if err != nil {
		panic(err.Error())
	}
	s := []currency{}
	for _, val := range split {
		s = append(s, currency(val.Amount()))
	}
	return s
}

func (c currency) Mul(multiple int64) currency {
	val := c.Value().Multiply(multiple)
	return currency(val.Amount())
}

type Payload struct {
	People  []People      `json:"people"`
	Shared  []SharedItems `json:"sharedItems"`
	TipPaid currency      `json:"tipPaid"`
	TaxPaid currency      `json:"taxPaid"`
}

type People struct {
	Name      string `json:"name"`
	Purchases []Item `json:"items"`
}

type Item struct {
	Name     string   `json:"itemName"`
	Price    currency `json:"price"`
	Quantity int64    `json:"quantity"`
	Total    currency `json:"total"`
}

type SharedItems struct {
	People []struct {
		Name string `json:"name"`
	} `json:"people"`
	Purchases []Item `json:"items"`
}

type Receipt struct {
	Name        string   `json:"name"`
	Items       []Item   `json:"items"`
	SharedItems []Item   `json:"sharedItems"`
	ItemSum     currency `json:"itemSum"`
	Tax         currency `json:"tax"`
	Tip         currency `json:"tip"`
	Total       currency `json:"total"`
}

type Result struct {
	Receipt   map[string]Receipt
	BillTotal currency
	Data      Payload

	// Required so that items are calculated accordingly
	taxPercentage decimal.Decimal
	tipPercentage decimal.Decimal
}

func (r *Result) processIndividual() {
	for _, person := range r.Data.People {
		for _, item := range person.Purchases {
			item.Total = item.Price.Mul(item.Quantity)
			receipt := r.Receipt[person.Name]
			receipt.ItemSum = receipt.ItemSum.Add(item.Total)
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
			item.Total = item.Price.Mul(item.Quantity)
			splitPrice := item.Total.Split(len(group.People))
			fmt.Println(splitPrice)
			for i, people := range group.People {
				receipt := r.Receipt[people.Name]
				receipt.SharedItems = append(receipt.SharedItems, Item{
					Name:     item.Name,
					Price:    item.Price,
					Quantity: item.Quantity,
					Total:    splitPrice[i],
				})
				receipt.ItemSum = receipt.ItemSum.Add(splitPrice[i])
				r.Receipt[people.Name] = receipt
			}
		}

	}
}

func (r *Result) postProcess() {
	for key, val := range r.Receipt {
		receipt := r.Receipt[key]
		receipt.Name = key
		receipt.ItemSum = val.ItemSum
		r.Receipt[key] = receipt

	}

	shares := []Receipt{}
	for _, val := range r.Receipt {
		shares = append(shares, val)
	}
	// each index represents shares[i] item sum
	itemSums := []int{}
	for _, val := range shares {
		itemSums = append(itemSums, int(val.ItemSum))
	}
	taxShares, err := r.Data.TaxPaid.Value().Allocate(itemSums...)
	if err != nil {
		panic(err.Error())
	}
	tipShares, err := r.Data.TipPaid.Value().Allocate(itemSums...)
	if err != nil {
		panic(err.Error())
	}
	for i, val := range shares {
		tax := currency(taxShares[i].Amount())
		tip := currency(tipShares[i].Amount())
		receipt := r.Receipt[val.Name]
		receipt.Tax = tax
		receipt.Tip = tip
		receipt.Total = tax.Add(tip).Add(receipt.ItemSum)

		r.BillTotal = r.BillTotal.Add(receipt.Total)
		r.Receipt[val.Name] = receipt

	}
}

func (r *Result) Process() string {

	r.Receipt = make(map[string]Receipt)
	r.BillTotal = 0

	// pre process
	r.processIndividual()
	r.processSplit()

	// post process
	r.postProcess()

	output, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		fmt.Printf("Error marshalling map: %v\n", err.Error())
	}

	return string(output)
}

func (c currency) MarshaJSON() ([]byte, error) {
	return json.Marshal(c.Value().AsMajorUnits())
}

func (mapping *Result) MarshalJSON() ([]byte, error) {
	decimal.MarshalJSONWithoutQuotes = true
	buffer := bytes.NewBufferString("{\"people\":[")
	length := len(mapping.Receipt)
	count := 0
	var itemTotal, taxTotal, tipTotal currency
	for _, val := range mapping.Receipt {
		taxTotal = taxTotal.Add(val.Tax)
		tipTotal = tipTotal.Add(val.Tip)
		itemTotal = itemTotal.Add(val.ItemSum)
	}
	for _, val := range mapping.Receipt {
		jsonVal, _ := json.Marshal(val)
		buffer.WriteString(string(jsonVal))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("],")
	buffer.WriteString(fmt.Sprintf("\"itemTotal\": %v,\n", itemTotal))
	buffer.WriteString(fmt.Sprintf("\"taxTotal\": %v,\n", taxTotal))
	buffer.WriteString(fmt.Sprintf("\"tipTotal\": %v,\n", tipTotal))
	buffer.WriteString(fmt.Sprintf("\"billTotal\": %v}", mapping.BillTotal))
	return buffer.Bytes(), nil
}
