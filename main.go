package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
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
		Key    string
	}

	err := json.NewDecoder(r.Body).Decode(&Data)
	if err != nil {
		fmt.Println("ERR: Body not in expected format")
		return
	}

	bytes, err := base64.StdEncoding.DecodeString(Data.Key)
	if err != nil {
		fmt.Println(err)
	}
	pem, _ := pem.Decode(bytes)
	if pem == nil {
		fmt.Println("Error decoding pem block")
	}
	key, err := x509.ParsePKCS1PublicKey(pem.Bytes)
	if err != nil {
		fmt.Println(err)
	}

	// Create a cryptographically random password
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	charsLen := big.NewInt(int64(len(chars)))
	length := 16
	var b strings.Builder
	for i := 0; i < length; i++ {
		pos, _ := rand.Int(rand.Reader, charsLen)
		b.WriteRune(chars[pos.Int64()])
	}
	pwd := b.String()
	encPwdBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, []byte(pwd), []byte("pwd"))
	if err != nil {
		fmt.Println("Encryption error", err)
	}
	encPwdStr := base64.StdEncoding.EncodeToString(encPwdBytes)

	// Create a lock with the given data and generated password
	l := Lock{Data.User, Data.Name, Data.Serial, encPwdStr}
	registeredLocks = append(registeredLocks, l)
	json.NewEncoder(w).Encode(l)
}

// Request access to a lock
func access(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	serial := params["serial"]

	user := r.FormValue("user")

	lock, err := get_lock(serial)
	if err == nil {
		if lock.Owner == user {
			json.NewEncoder(w).Encode(lock)
		}
	}
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

	params := mux.Vars(r)
	serial := params["serial"]

	lock, err := get_lock(serial)

	if err == nil {
		temp := lock.Password
		lock.Password = ""
		json.NewEncoder(w).Encode(lock)
		lock.Password = temp
	}
}

func get_lock(serial string) (*Lock, error) {
	for _, lock := range registeredLocks {
		if lock.Serial == serial {
			return &lock, nil
		}
	}

	return &Lock{}, errors.New("No lock exists by that serial")
}

func main() {
	registeredLocks = make([]Lock, 0)

	r := mux.NewRouter()
	r.HandleFunc("/", root).Methods(http.MethodGet)
	r.HandleFunc("/locks", locks).Queries("user", "{user}").Methods(http.MethodGet)
	r.HandleFunc("/locks", register).Methods(http.MethodPost)
	r.HandleFunc("/locks/{serial}/access", access).Queries("user", "{user}").Methods(http.MethodGet)
	r.HandleFunc("/locks/{serial}", lock_by_serial).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServeTLS(":8080", "https-server.crt", "https-server.key", r))
	//log.Fatal(http.ListenAndServe(":8080", r))
}
