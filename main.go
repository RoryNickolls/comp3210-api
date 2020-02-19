package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var registeredLocks map[string]Lock

type Lock struct {
	ID              int
	EncryptedSerial string
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Hello World`))
}

func register(w http.ResponseWriter, r *http.Request) {
	var Data struct {
		Serial string
	}

	err := json.NewDecoder(r.Body).Decode(&Data)
	if err != nil {
		fmt.Println("ERR: Body not in expected format")
		return
	}

	id := len(registeredLocks) + 1
	l := Lock{id, Data.Serial + "ENC"}
	registeredLocks[strconv.Itoa(id)] = l
	json.NewEncoder(w).Encode(l)
}

func access(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	l := registeredLocks[params["id"]]
	json.NewEncoder(w).Encode(l)
}

func locks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(registeredLocks)
}

func main() {
	registeredLocks = make(map[string]Lock)

	r := mux.NewRouter()
	r.HandleFunc("/", root).Methods(http.MethodGet)
	r.HandleFunc("/locks", locks).Methods(http.MethodGet)
	r.HandleFunc("/locks/{id}/access", access).Methods(http.MethodGet)
	r.HandleFunc("/locks", register).Methods(http.MethodPost)
	log.Fatal(http.ListenAndServe(":8080", r))
}
