package main

import (
	"fmt"
	"html/template"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/eliasffyksen/GTML/gtml"
)

type CustomHTML struct{}

var _ gtml.HTMLer = CustomHTML{}

func (ch CustomHTML) HTML() (template.HTML, error) {
	return template.HTML("<span>this is a test</span>"), nil
}

type Index struct {
	ShoppingLists gtml.SearchLink[ShoppingListSearch]
	Products      []*Product
}

type ShoppingListSearch struct {
	Name string `gtml:"search"`
}

type ShoppingList struct {
	Name        string
	Description string
	Items       []Item
	CustomHtml  CustomHTML `gtml:"table-hide"`
}

// Link implements gtml.Linker.
func (s *ShoppingList) Link() any {
	return ShoppingListLink{
		ShoppingList: s.Name,
	}
}

var _ gtml.Linker = &ShoppingList{}

type Item struct {
	Product  *Product
	Quantity int
}

type ShoppingListLink struct {
	ShoppingList string
}

type Product struct {
	Name        string
	Description string
	Nutrients   []Nutrient
}

func (p *Product) Link() any {
	return ProductLink{
		p.Name,
	}
}

func (p *Product) String() string {
	return p.Name
}

type Nutrient struct {
	Name   string
	Amount float64
}

type ProductLink struct {
	Product string
}

var _ gtml.Linker = &Product{}

var products = map[string]*Product{
	"Jarlsberg": {
		Name:        "Jarlsberg",
		Description: "A mild nutty Alipine-style cheese from Norway",
		Nutrients: []Nutrient{
			{"Cheese", 50},
			{"Sugar", 50},
		},
	},
	"Kvikk-Lunsj": {
		Name:        "Kvikk-Lunsj",
		Description: "The supirior version of KitKat",
	},
}

var shoppingLists = map[string]*ShoppingList{
	"My Shopping List": {
		Name:        "My Shopping List",
		Description: "Stuff I buy before and after work",
		Items: []Item{
			{Product: products["Jarlsberg"], Quantity: 10},
			{Product: products["Kvikk-Lunsj"], Quantity: 100},
		},
	},
	"Not My Shopping List": {
		Name:        "Not My Shopping List",
		Description: "Stuff I buy don't buybefore and after work",
		Items: []Item{
			{Product: products["Jarlsberg"], Quantity: -10},
			{Product: products["Kvikk-Lunsj"], Quantity: -100},
		},
	},
}

func ProductGetter(link ProductLink) (Product, error) {
	product, ok := products[link.Product]
	if !ok {
		return Product{}, fmt.Errorf("404 - Product %s not found", link.Product)
	}

	return *product, nil
}

func ShoppingLinkGetter(link ShoppingListLink) (ShoppingList, error) {
	list, ok := shoppingLists[link.ShoppingList]
	if !ok {
		return ShoppingList{}, fmt.Errorf("404 - Shopping list %s not found", link.ShoppingList)
	}

	return *list, nil
}

func ShoppingListSearcher(link ShoppingListSearch) ([]*ShoppingList, error) {
	searchTerm := strings.ToLower(link.Name)
	results := make([]*ShoppingList, 0)

	for item := range maps.Values(shoppingLists) {
		name := strings.ToLower(item.Name)

		if strings.Contains(name, searchTerm) {
			results = append(results, item)
		}
	}

	return results, nil
}

func IndexGetter(link struct{}) (Index, error) {
	return Index{
		ShoppingLists: gtml.NewSearchLink(ShoppingListSearch{}),
		Products:      slices.Collect(maps.Values(products)),
	}, nil
}

func main() {
	router := gtml.NewRouter()

	if err := gtml.Get(router, IndexGetter); err != nil {
		panic(err)
	}

	if err := gtml.Get(router, ProductGetter); err != nil {
		panic(err)
	}

	if err := gtml.Get(router, ShoppingLinkGetter); err != nil {
		panic(err)
	}

	if err := gtml.Search(router, ShoppingListSearcher); err != nil {
		panic(err)
	}

	err := http.ListenAndServe(":8080", router.Mux)
	if err != nil {
		panic(err)
	}
}
