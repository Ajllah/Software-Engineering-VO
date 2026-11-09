package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ski-verleih/internal/store"
)

type API struct {
	store *store.Store
}

func NewAPI(dataStore *store.Store) *API {
	return &API{store: dataStore}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/dashboard" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, a.store.Dashboard())
		return
	}

	switch {
	case r.URL.Path == "/api/config":
		a.handleConfig(w, r)
	case r.URL.Path == "/api/reservations":
		a.handleReservations(w, r)
	case r.URL.Path == "/api/skis":
		a.handleSkis(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/skis/"):
		a.handleSki(w, r)
	case r.URL.Path == "/api/customers":
		a.handleCustomers(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/customers/"):
		a.handleCustomer(w, r)
	case r.URL.Path == "/api/rentals":
		a.handleRentals(w, r)
	case r.URL.Path == "/api/rentals/return":
		a.handleReturn(w, r)
	default:
		writeError(w, http.StatusNotFound, "Endpoint nicht gefunden")
	}
}

func (a *API) handleSkis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListSkis())
	case http.MethodPost:
		var request struct {
			Type string `json:"type"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		ski, err := a.store.CreateSki(request.Type)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, ski)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleSki(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r.URL.Path, "/api/skis/")
	if !ok {
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}

	var request struct {
		Type string `json:"type"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	ski, err := a.store.UpdateSki(id, request.Type)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ski)
}

func (a *API) handleCustomers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListCustomers())
	case http.MethodPost:
		var request customerRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		customer, err := a.store.CreateCustomer(request.Name, request.DocumentType, request.DocumentNumber)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, customer)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r.URL.Path, "/api/customers/")
	if !ok {
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}

	var request customerRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	customer, err := a.store.UpdateCustomer(id, request.Name, request.DocumentType, request.DocumentNumber)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, customer)
}

func (a *API) handleRentals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListRentals())
	case http.MethodPost:
		a.handleCheckout(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleReservations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var request struct {
		SkiID      int    `json:"skiId"`
		CustomerID int    `json:"customerId"`
		StartDate  string `json:"startDate"`
		Days       int    `json:"days"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	startDate, err := time.ParseInLocation(time.DateOnly, request.StartDate, time.Local)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Verleihbeginn muss ein gültiges Datum im Format YYYY-MM-DD sein")
		return
	}
	rental, err := a.store.Reserve(request.SkiID, request.CustomerID, request.Days, startDate, time.Now())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rental)
}

func (a *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]int64{"dailyRateCents": a.store.DailyRateCents()})
	case http.MethodPut:
		var request struct {
			DailyRateCents *int64 `json:"dailyRateCents"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.DailyRateCents == nil {
			writeError(w, http.StatusBadRequest, "dailyRateCents ist erforderlich")
			return
		}
		if err := a.store.SetDailyRateCents(*request.DailyRateCents); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]int64{"dailyRateCents": a.store.DailyRateCents()})
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleReturn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var request struct {
		SkiID int `json:"skiId"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	rental, err := a.store.ReturnSki(request.SkiID, time.Now())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rental)
}

func (a *API) handleCheckout(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SkiID      int `json:"skiId"`
		CustomerID int `json:"customerId"`
		Days       int `json:"days"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	rental, err := a.store.Checkout(request.SkiID, request.CustomerID, request.Days, time.Now())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rental)
}

type customerRequest struct {
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "Ungültige JSON-Eingabe: "+err.Error())
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "Es ist nur ein JSON-Objekt pro Anfrage erlaubt")
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, path, prefix string) (int, bool) {
	part := strings.TrimPrefix(path, prefix)
	if part == "" || strings.Contains(part, "/") {
		writeError(w, http.StatusNotFound, "Eintrag nicht gefunden")
		return 0, false
	}
	id, err := strconv.Atoi(part)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "ID muss eine positive Zahl sein")
		return 0, false
	}
	return id, true
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrAlreadyRented), errors.Is(err, store.ErrNotRented), errors.Is(err, store.ErrUnavailable):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Interner Fehler: %v", err))
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "HTTP-Methode nicht erlaubt")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "Antwort konnte nicht geschrieben werden", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
