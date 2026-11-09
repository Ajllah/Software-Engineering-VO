package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"fahrrad-verleih/internal/store"
)

func TestCreateResourcesBookingConflictAndReturn(t *testing.T) {
	dataStore, err := store.NewWithHourlyRateCents(500)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewAPI(dataStore))
	defer server.Close()

	bike := postJSON(t, server.URL+"/api/bikes", map[string]string{"type": "Citybike"}, http.StatusCreated)
	customer := postJSON(t, server.URL+"/api/customers", map[string]string{
		"name": "Ada Beispiel", "documentType": "Personalausweis", "documentNumber": "A123",
	}, http.StatusCreated)

	var bikeData struct {
		ID int `json:"id"`
	}
	var customerData struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(bike, &bikeData); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(customer, &customerData); err != nil {
		t.Fatal(err)
	}
	start := "2099-05-10T10:00"
	first := postJSON(t, server.URL+"/api/bookings", map[string]any{
		"bikeId": bikeData.ID, "customerId": customerData.ID,
		"startTime": start, "endTime": "2099-05-10T12:00",
	}, http.StatusCreated)
	var booking struct {
		ID              int   `json:"id"`
		TotalPriceCents int64 `json:"totalPriceCents"`
	}
	if err := json.Unmarshal(first, &booking); err != nil {
		t.Fatal(err)
	}
	if booking.TotalPriceCents != 1000 {
		t.Fatalf("total price = %d cents, want 1000", booking.TotalPriceCents)
	}
	postJSON(t, server.URL+"/api/bookings", map[string]any{
		"bikeId": bikeData.ID, "customerId": customerData.ID,
		"startTime": "2099-05-10T11:00", "endTime": "2099-05-10T13:00",
	}, http.StatusConflict)

	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/bookings/"+jsonNumber(booking.ID)+"/return", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("return for future booking status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestAPIRejectsMissingCustomerFieldsAndUnknownJSONFields(t *testing.T) {
	server := httptest.NewServer(NewAPI(store.New()))
	defer server.Close()
	postJSON(t, server.URL+"/api/customers", map[string]string{"name": "Ada"}, http.StatusBadRequest)

	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/bikes", bytes.NewBufferString(`{"type":"Citybike","id":9}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestImmediateRentalCanBeReturned(t *testing.T) {
	server := httptest.NewServer(NewAPI(store.New()))
	defer server.Close()
	bike := postJSON(t, server.URL+"/api/bikes", map[string]string{"type": "Citybike"}, http.StatusCreated)
	customer := postJSON(t, server.URL+"/api/customers", map[string]string{
		"name": "Ada Beispiel", "documentType": "Personalausweis", "documentNumber": "A123",
	}, http.StatusCreated)
	var bikeData struct {
		ID int `json:"id"`
	}
	var customerData struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(bike, &bikeData); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(customer, &customerData); err != nil {
		t.Fatal(err)
	}
	body := postJSON(t, server.URL+"/api/bookings", map[string]any{
		"bikeId": bikeData.ID, "customerId": customerData.ID,
		"startNow": true, "endTime": time.Now().Add(time.Hour).Format("2006-01-02T15:04"),
	}, http.StatusCreated)
	var booking struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &booking); err != nil {
		t.Fatal(err)
	}
	postJSON(t, server.URL+"/api/bookings/"+strconv.Itoa(booking.ID)+"/return", map[string]string{}, http.StatusOK)
}

func postJSON(t *testing.T, url string, body any, status int) []byte {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(url, "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != status {
		t.Fatalf("POST %s status = %d, want %d", url, response.StatusCode, status)
	}
	var result json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

func jsonNumber(value int) string {
	return strconv.Itoa(value)
}
