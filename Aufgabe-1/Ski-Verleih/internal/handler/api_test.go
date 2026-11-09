package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ski-verleih/internal/store"
)

func TestAPIWorkflow(t *testing.T) {
	api := NewAPI(store.New())

	ski := performJSONRequest(t, api, http.MethodPost, "/api/skis", `{"type":"Touring"}`, http.StatusCreated)
	if ski["id"] != float64(1) {
		t.Fatalf("Ski-ID = %v; erwartet 1", ski["id"])
	}
	customer := performJSONRequest(t, api, http.MethodPost, "/api/customers", `{"name":"Alex Beispiel","documentType":"Reisepass","documentNumber":"X12345"}`, http.StatusCreated)
	if customer["id"] != float64(1) {
		t.Fatalf("Kunden-ID = %v; erwartet 1", customer["id"])
	}

	rental := performJSONRequest(t, api, http.MethodPost, "/api/rentals", `{"skiId":1,"customerId":1,"days":2}`, http.StatusCreated)
	if rental["days"] != float64(2) {
		t.Fatalf("Ausleihdauer = %v; erwartet 2", rental["days"])
	}
	performJSONRequest(t, api, http.MethodPost, "/api/rentals", `{"skiId":1,"customerId":1,"days":1}`, http.StatusConflict)

	dashboard := performJSONRequest(t, api, http.MethodGet, "/api/dashboard", "", http.StatusOK)
	if dashboard["availableSkis"] != float64(0) || dashboard["rentedSkis"] != float64(1) {
		t.Fatalf("Dashboard nach Ausleihe ist falsch: %#v", dashboard)
	}

	performJSONRequest(t, api, http.MethodPost, "/api/rentals/return", `{"skiId":1}`, http.StatusOK)
	dashboard = performJSONRequest(t, api, http.MethodGet, "/api/dashboard", "", http.StatusOK)
	if dashboard["availableSkis"] != float64(1) || dashboard["rentedSkis"] != float64(0) {
		t.Fatalf("Dashboard nach Rückgabe ist falsch: %#v", dashboard)
	}
}

func TestAPIRejectsInvalidInputAndUnknownFields(t *testing.T) {
	api := NewAPI(store.New())
	performJSONRequest(t, api, http.MethodPost, "/api/skis", `{"type":" "}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodPost, "/api/skis", `{"id":42,"type":"Touring"}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodPost, "/api/customers", `{"name":"Alex","documentType":"Pass","documentNumber":""}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodPost, "/api/rentals", `{"skiId":1,"customerId":1,"days":0}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodGet, "/api/unknown", "", http.StatusNotFound)
	performJSONRequest(t, api, http.MethodDelete, "/api/skis", "", http.StatusMethodNotAllowed)
}

func TestReservationAndDailyRateAPI(t *testing.T) {
	api := NewAPI(store.New())
	performJSONRequest(t, api, http.MethodPost, "/api/skis", `{"type":"Touring"}`, http.StatusCreated)
	performJSONRequest(t, api, http.MethodPost, "/api/customers", `{"name":"Alex Beispiel","documentType":"Reisepass","documentNumber":"X12345"}`, http.StatusCreated)

	config := performJSONRequest(t, api, http.MethodGet, "/api/config", "", http.StatusOK)
	if config["dailyRateCents"] != float64(store.DefaultDailyRateCents) {
		t.Fatalf("Starttarif = %v Cent; erwartet %d", config["dailyRateCents"], store.DefaultDailyRateCents)
	}
	config = performJSONRequest(t, api, http.MethodPut, "/api/config", `{"dailyRateCents":2350}`, http.StatusOK)
	if config["dailyRateCents"] != float64(2350) {
		t.Fatalf("gespeicherter Tagestarif = %v Cent; erwartet 2350", config["dailyRateCents"])
	}
	performJSONRequest(t, api, http.MethodPut, "/api/config", `{}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodPut, "/api/config", `{"dailyRateCents":-1}`, http.StatusBadRequest)

	startDate := time.Now().In(time.Local).AddDate(0, 0, 2).Format(time.DateOnly)
	body, err := json.Marshal(map[string]any{"skiId": 1, "customerId": 1, "startDate": startDate, "days": 4})
	if err != nil {
		t.Fatal(err)
	}
	reservation := performJSONRequest(t, api, http.MethodPost, "/api/reservations", string(body), http.StatusCreated)
	if reservation["dailyRateCents"] != float64(2350) || reservation["totalPriceCents"] != float64(9400) {
		t.Fatalf("Reservierung enthält falschen Tarif oder Gesamtpreis: %#v", reservation)
	}

	overlapStart := time.Now().In(time.Local).AddDate(0, 0, 3).Format(time.DateOnly)
	overlapBody, err := json.Marshal(map[string]any{"skiId": 1, "customerId": 1, "startDate": overlapStart, "days": 1})
	if err != nil {
		t.Fatal(err)
	}
	performJSONRequest(t, api, http.MethodPost, "/api/reservations", string(overlapBody), http.StatusConflict)
	performJSONRequest(t, api, http.MethodPost, "/api/reservations", `{"skiId":1,"customerId":1,"startDate":"not-a-date","days":1}`, http.StatusBadRequest)
	performJSONRequest(t, api, http.MethodPost, "/api/reservations", `{"skiId":1,"customerId":1,"startDate":"2020-01-01","days":1}`, http.StatusBadRequest)
}

func performJSONRequest(t *testing.T, api http.Handler, method, path, body string, wantStatus int) map[string]any {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s: Status = %d, erwartet %d; Antwort: %s", method, path, response.Code, wantStatus, response.Body.String())
	}
	if response.Body.Len() == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("%s %s: ungültige JSON-Antwort: %v", method, path, err)
	}
	return result
}
