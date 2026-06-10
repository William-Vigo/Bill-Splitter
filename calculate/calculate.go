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
	fmt.Println(c)
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

func (c currency) Mul(multiple decimal.Decimal) currency {
	price := decimal.NewFromFloat(float64(c.Value().Amount()))
	val := multiple.Mul(price).IntPart()
	return currency(val)
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
	Name  string   `json:"itemName"`
	Price currency `json:"price"`
}

type SharedItems struct {
	People []struct {
		Name string `json:"name"`
	} `json:"people"`
	Purchases []Item `json:"items"`
}

type Receipt struct {
	Name        string   `json:"name"`
	Items       []Item   `json: "items"`
	SharedItems []Item   `json: "sharedItems"`
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
			receipt := r.Receipt[person.Name]
			receipt.ItemSum = receipt.ItemSum.Add(item.Price)
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
			splitPrice := item.Price.Split(len(group.People))
			for i, people := range group.People {
				receipt := r.Receipt[people.Name]
				receipt.SharedItems = append(receipt.SharedItems, Item{
					Name:  item.Name,
					Price: splitPrice[i],
				})
				receipt.ItemSum = receipt.ItemSum.Add(splitPrice[i])
				r.Receipt[people.Name] = receipt
			}
		}

	}
}
func (r *Result) calculatePercentages() {
	var total currency
	for _, val := range r.Receipt {
		total = total.Add(val.ItemSum)
	}
	r.taxPercentage = decimal.NewFromFloat(float64(r.Data.TaxPaid)).
		Div(decimal.NewFromFloat(float64(total)))
	r.tipPercentage = decimal.NewFromFloat(float64(r.Data.TipPaid)).
		Div(decimal.NewFromFloat(float64(total)))
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
	r.BillTotal = 0

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
