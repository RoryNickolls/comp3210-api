package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var registeredLocks []Lock

type Lock struct {
	Owner    string
	Name     string
	Serial   string
	Password string
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Register a lock to this account
func register(w http.ResponseWriter, r *http.Request) {
	var Data struct {
		User   string
		Name   string
		Serial string
	}

	err := json.NewDecoder(r.Body).Decode(&Data)
	if err != nil {
		fmt.Println("ERR: Body not in expected format")
		return
	}

	// Create a cryptographically random password
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	charsLen := big.NewInt(int64(len(chars)))
	length := 32
	var b strings.Builder
	for i := 0; i < length; i++ {
		pos, _ := rand.Int(rand.Reader, charsLen)
		b.WriteRune(chars[pos.Int64()])
	}

	// Create a lock with the given data and generated password
	l := Lock{Data.User, Data.Name, Data.Serial, b.String()}
	registeredLocks = append(registeredLocks, l)
	json.NewEncoder(w).Encode(l)
}

// Request access to a lock
func access(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	i, _ := strconv.Atoi(params["id"])
	l := registeredLocks[i]
	json.NewEncoder(w).Encode(l)
}

// Get the list of registered locks
func locks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := r.FormValue("user")

	var userLocks []Lock
	for _, lock := range registeredLocks {
		if lock.Owner == user {
			userLocks = append(userLocks, lock)
		}
	}
	json.NewEncoder(w).Encode(userLocks)
}

func lock_by_serial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	serial := r.FormValue("serial")

	for _, lock := range registeredLocks {
		if lock.Serial == serial {
			temp := lock.Password
			lock.Password = ""
			json.NewEncoder(w).Encode(lock)
			lock.Password = temp
			return
		}
	}
}

func main() {
	registeredLocks = make([]Lock, 0)

	r := mux.NewRouter()
	r.HandleFunc("/", root).Methods(http.MethodGet)
	r.HandleFunc("/locks", locks).Queries("user", "{user}").Methods(http.MethodGet)
	r.HandleFunc("/locks/{id}/access", access).Methods(http.MethodGet)
	r.HandleFunc("/locks", register).Methods(http.MethodPost)
	r.HandleFunc("/lock", lock_by_serial).Queries("serial", "{serial}").Methods(http.MethodGet)
	//log.Fatal(http.ListenAndServeTLS(":8080", "cert.crt", "key.key", r))
	log.Fatal(http.ListenAndServe(":8080", r))
}
