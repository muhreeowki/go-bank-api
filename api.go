package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// APIServer is a struct that represents an API server
type APIServer struct {
	listenAddr string
	store      Storage
}

// APIFunc is a http.HanderFunc that returns an error
type APIFunc func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Error string
	Code  int
}

// NewAPIServer creates a new APIServer with the given listen address
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
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
	router.HandleFunc("/accounts", MakeHTTPHandlerFunc(s.handleGetAccounts))

	log.Println("JSON API is running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
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

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, NewAccount("Muriuki", "Muchiri"))
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	return nil
}
