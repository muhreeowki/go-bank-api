package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// APIServer is a struct that represents an API server
type APIServer struct {
	store      Storage
	listenAddr string
}

// APIFunc is a http.HanderFunc that returns an error
type APIFunc func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// NewAPIServer creates a new APIServer with the given listen address
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		store:      store,
		listenAddr: listenAddr,
	}
}

// WriteJSON writes a JSON response with the given status code and object
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	// Set the status code and content type for the response
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	// Encode the v object to JSON
	err := json.NewEncoder(w).Encode(v)
	return err
}

// MakeHTTPHandlerFunc wraps an APIFunc to handle errors and write JSON responses
func MakeHTTPHandlerFunc(fn APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			// handle error here
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", MakeHTTPHandlerFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", MakeHTTPHandlerFunc(s.handleAccountByID))
	router.HandleFunc("/transfer", MakeHTTPHandlerFunc(s.handleTransfer))

	log.Println("JSON API is running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccounts(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return fmt.Errorf("method not allowed: %s", r.Method)
	}
}

func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccountByID(w, r)
	case "DELETE":
		return s.handleDeleteAccountByID(w, r)
	default:
		return fmt.Errorf("method not allowed: %s", r.Method)
	}
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	defer r.Body.Close()

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "successfully deleted account")
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return fmt.Errorf("bad Request: %s", err)
	}
	defer r.Body.Close()

	// msg := fmt.Sprintf("successfully transfered amount %s to account %s", transferReq.Amount, transferReq.ToAccount)
	return WriteJSON(w, http.StatusOK, transferReq)
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
