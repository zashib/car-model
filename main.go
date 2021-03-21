package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	StatusOnTheWay          = "on the way"
	StatusInStock           = "in stock"
	StatusSoldOut           = "sold out"
	StatusWithdrawnFromSale = "withdrawn from sale"
)

type CarModel struct {
	Brand   string `json:"brand"`
	Model   string `json:"model"`
	Price   uint   `json:"price"`
	Status  string `json:"status"`
	Mileage uint   `json:"mileage"`
}

type carHandlers struct {
	sync.Mutex
	store map[string]CarModel
}

func (c *carHandlers) insert(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()

	var car CarModel

	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	id := uuid.New().String()
	c.store[id] = car

	w.WriteHeader(http.StatusCreated)
}

func (c *carHandlers) get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	if v, ok := c.store[id]; ok {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.Header().Add("content-type", "application/json")
		w.Write(jsonBytes)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *carHandlers) list(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()

	var cars []CarModel
	for _, car := range c.store {
		cars = append(cars, car)
	}

	jsonBytes, err := json.Marshal(cars)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jsonBytes)
}

func (c *carHandlers) update(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()

	var car CarModel

	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	params := mux.Vars(r)
	id := params["id"]

	if _, ok := c.store[id]; ok {
		c.store[id] = car

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *carHandlers) delete(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()

	params := mux.Vars(r)
	id := params["id"]
	delete(c.store, id)
	w.WriteHeader(http.StatusOK)
}

func main() {
	handler := carHandlers{store: map[string]CarModel{
		uuid.New().String(): {
			Brand:   "nissan",
			Model:   "almera",
			Price:   20000,
			Status:  StatusInStock,
			Mileage: 30000,
		},
	}}
	r := mux.NewRouter()
	r.HandleFunc("/cars", handler.insert).Methods(http.MethodPost)
	r.HandleFunc("/cars/{id}", handler.get).Methods(http.MethodGet)
	r.HandleFunc("/cars", handler.list).Methods(http.MethodGet)
	r.HandleFunc("/cars/{id}", handler.update).Methods(http.MethodPut)
	r.HandleFunc("/cars/{id}", handler.delete).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8000", r))
}
