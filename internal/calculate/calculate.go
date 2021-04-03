package calculate

import (
	"fmt"
	"sync"
)

type Payload struct {
	Group  []People      `json:"people"`
	Shared []SharedItems `json:"sharedItems"`
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

//TODO: calculate tax & tip
func Process(data Payload) {
	moneyOwed := make(map[string]float64)

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
					moneyOwed[person.Name] += purchases.Price
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
					for _, purchaces := range group.Purchases {
						total += purchaces.Price
					}

					moneyDue := total / float64(splitSize)
					//TODO round money to 2 decimal places

					for _, people := range group.People {
						lock.Lock()
						moneyOwed[people.Name] += moneyDue
						lock.Unlock()
					}

				}(group)

			}
		}()
	}

	wg.Wait()
	fmt.Println(moneyOwed)
}
