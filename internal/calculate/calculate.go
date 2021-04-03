package calculate

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
