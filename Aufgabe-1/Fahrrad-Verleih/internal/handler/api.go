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

	"fahrrad-verleih/internal/store"
)

type API struct {
	store *store.Store
}

func NewAPI(dataStore *store.Store) *API {
	return &API{store: dataStore}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/dashboard":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		writeJSON(w, http.StatusOK, a.store.Dashboard(time.Now()))
	case r.URL.Path == "/api/config":
		a.handleConfig(w, r)
	case r.URL.Path == "/api/bikes":
		a.handleBikes(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/bikes/"):
		a.handleBike(w, r)
	case r.URL.Path == "/api/customers":
		a.handleCustomers(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/customers/"):
		a.handleCustomer(w, r)
	case r.URL.Path == "/api/bookings":
		a.handleBookings(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/bookings/"):
		a.handleBooking(w, r)
	default:
		writeError(w, http.StatusNotFound, "Endpoint nicht gefunden")
	}
}

func (a *API) handleBikes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListBikes())
	case http.MethodPost:
		var request struct {
			Type string `json:"type"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		bike, err := a.store.CreateBike(request.Type)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, bike)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleBike(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r.URL.Path, "/api/bikes/")
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
	bike, err := a.store.UpdateBike(id, request.Type)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bike)
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

func (a *API) handleBookings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.store.ListBookings())
	case http.MethodPost:
		var request struct {
			BikeID     int    `json:"bikeId"`
			CustomerID int    `json:"customerId"`
			StartNow   bool   `json:"startNow"`
			StartTime  string `json:"startTime"`
			EndTime    string `json:"endTime"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		now := time.Now()
		start := now
		var err error
		if !request.StartNow {
			start, err = parseDateTime(request.StartTime)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Buchungsbeginn muss ein gültiges Datum mit Uhrzeit sein")
				return
			}
		}
		end, err := parseDateTime(request.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Buchungsende muss ein gültiges Datum mit Uhrzeit sein")
			return
		}
		booking, err := a.store.CreateBooking(request.BikeID, request.CustomerID, start, end, now)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, booking)
	default:
		methodNotAllowed(w)
	}
}

func (a *API) handleBooking(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/return") || r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	id, ok := pathID(w, strings.TrimSuffix(r.URL.Path, "/return"), "/api/bookings/")
	if !ok {
		return
	}
	booking, err := a.store.ReturnBooking(id, time.Now())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booking)
}

func (a *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]int64{"hourlyRateCents": a.store.HourlyRateCents()})
	case http.MethodPut:
		var request struct {
			HourlyRateCents *int64 `json:"hourlyRateCents"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.HourlyRateCents == nil {
			writeError(w, http.StatusBadRequest, "hourlyRateCents ist erforderlich")
			return
		}
		if err := a.store.SetHourlyRateCents(*request.HourlyRateCents); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]int64{"hourlyRateCents": a.store.HourlyRateCents()})
	default:
		methodNotAllowed(w)
	}
}

type customerRequest struct {
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
}

func parseDateTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.ParseInLocation("2006-01-02T15:04", value, time.Local)
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
	case errors.Is(err, store.ErrUnavailable):
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
