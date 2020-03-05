package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var registeredLocks []Lock

type Lock struct {
	Name     string
	Serial   string
	Password string
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func register(w http.ResponseWriter, r *http.Request) {
	var Data struct {
		Name   string
		Serial string
	}

	err := json.NewDecoder(r.Body).Decode(&Data)
	if err != nil {
		fmt.Println("ERR: Body not in expected format")
		return
	}

	l := Lock{Data.Name, Data.Serial, "open sesame"}
	registeredLocks = append(registeredLocks, l)
	json.NewEncoder(w).Encode(l)
}

func access(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	i, _ := strconv.Atoi(params["id"])
	l := registeredLocks[i]
	json.NewEncoder(w).Encode(l)
}

func locks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registeredLocks)
}

func main() {
	registeredLocks = make([]Lock, 0)

	r := mux.NewRouter()
	r.HandleFunc("/", root).Methods(http.MethodGet)
	r.HandleFunc("/locks", locks).Methods(http.MethodGet)
	r.HandleFunc("/locks/{id}/access", access).Methods(http.MethodGet)
	r.HandleFunc("/locks", register).Methods(http.MethodPost)
	log.Fatal(http.ListenAndServe(":8080", r))
}
