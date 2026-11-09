package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"fahrrad-verleih/internal/domain"
)

var (
	ErrNotFound    = errors.New("Eintrag nicht gefunden")
	ErrInvalid     = errors.New("ungültige Eingabe")
	ErrUnavailable = errors.New("Fahrrad ist im gewählten Zeitraum nicht verfügbar")
)

const DefaultHourlyRateCents int64 = 500

type Store struct {
	mu              sync.RWMutex
	bikes           map[int]domain.Bike
	customers       map[int]domain.Customer
	bookings        map[int]domain.Booking
	hourlyRateCents int64
	nextBikeID      int
	nextCustomerID  int
	nextBookingID   int
	dataFile        string
}

type persistedState struct {
	Bikes           []domain.Bike     `json:"bikes"`
	Customers       []domain.Customer `json:"customers"`
	Bookings        []domain.Booking  `json:"bookings"`
	HourlyRateCents int64             `json:"hourlyRateCents"`
	NextBikeID      int               `json:"nextBikeId"`
	NextCustomerID  int               `json:"nextCustomerId"`
	NextBookingID   int               `json:"nextBookingId"`
}

func New() *Store {
	return newStore(DefaultHourlyRateCents)
}

func NewWithHourlyRateCents(rate int64) (*Store, error) {
	if rate < 0 {
		return nil, fmt.Errorf("%w: Stundentarif darf nicht negativ sein", ErrInvalid)
	}
	return newStore(rate), nil
}

func newStore(rate int64) *Store {
	return &Store{
		bikes:           make(map[int]domain.Bike),
		customers:       make(map[int]domain.Customer),
		bookings:        make(map[int]domain.Booking),
		hourlyRateCents: rate,
		nextBikeID:      1,
		nextCustomerID:  1,
		nextBookingID:   1,
	}
}

func NewPersistent(dataFile string, initialHourlyRateCents int64) (*Store, error) {
	if strings.TrimSpace(dataFile) == "" {
		return nil, fmt.Errorf("JSON-Datenpfad darf nicht leer sein")
	}
	dataStore, err := NewWithHourlyRateCents(initialHourlyRateCents)
	if err != nil {
		return nil, err
	}
	dataStore.dataFile = dataFile

	content, err := os.ReadFile(dataFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return dataStore, nil
		}
		return nil, fmt.Errorf("JSON-Datendatei %q konnte nicht gelesen werden: %w", dataFile, err)
	}

	var state persistedState
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, fmt.Errorf("JSON-Datendatei %q ist ungültig: %w", dataFile, err)
	}
	if err := dataStore.loadState(state); err != nil {
		return nil, fmt.Errorf("JSON-Datendatei %q enthält ungültige Daten: %w", dataFile, err)
	}
	return dataStore, nil
}

func (s *Store) loadState(state persistedState) error {
	if state.HourlyRateCents < 0 {
		return fmt.Errorf("%w: negativer Stundentarif", ErrInvalid)
	}
	for _, bike := range state.Bikes {
		if bike.ID < 1 || strings.TrimSpace(bike.Type) == "" {
			return fmt.Errorf("%w: ungültiges Fahrrad in Datendatei", ErrInvalid)
		}
		if _, exists := s.bikes[bike.ID]; exists {
			return fmt.Errorf("%w: Fahrrad-ID %d mehrfach vorhanden", ErrInvalid, bike.ID)
		}
		s.bikes[bike.ID] = bike
		if bike.ID >= s.nextBikeID {
			s.nextBikeID = bike.ID + 1
		}
	}
	for _, customer := range state.Customers {
		if customer.ID < 1 || strings.TrimSpace(customer.Name) == "" ||
			strings.TrimSpace(customer.DocumentType) == "" || strings.TrimSpace(customer.DocumentNumber) == "" {
			return fmt.Errorf("%w: ungültiger Kunde in Datendatei", ErrInvalid)
		}
		if _, exists := s.customers[customer.ID]; exists {
			return fmt.Errorf("%w: Kunden-ID %d mehrfach vorhanden", ErrInvalid, customer.ID)
		}
		s.customers[customer.ID] = customer
		if customer.ID >= s.nextCustomerID {
			s.nextCustomerID = customer.ID + 1
		}
	}
	for _, booking := range state.Bookings {
		if booking.ID < 1 || booking.BikeID < 1 || booking.CustomerID < 1 ||
			booking.StartTime.IsZero() || !booking.EndTime.After(booking.StartTime) ||
			booking.HourlyRateCents < 0 || booking.TotalPriceCents < 0 {
			return fmt.Errorf("%w: ungültige Buchung in Datendatei", ErrInvalid)
		}
		if _, exists := s.bookings[booking.ID]; exists {
			return fmt.Errorf("%w: Buchungs-ID %d mehrfach vorhanden", ErrInvalid, booking.ID)
		}
		if _, exists := s.bikes[booking.BikeID]; !exists {
			return fmt.Errorf("%w: Fahrrad %d für Buchung %d fehlt", ErrInvalid, booking.BikeID, booking.ID)
		}
		if _, exists := s.customers[booking.CustomerID]; !exists {
			return fmt.Errorf("%w: Kunde %d für Buchung %d fehlt", ErrInvalid, booking.CustomerID, booking.ID)
		}
		s.bookings[booking.ID] = booking
		if booking.ID >= s.nextBookingID {
			s.nextBookingID = booking.ID + 1
		}
	}
	if state.NextBikeID > s.nextBikeID {
		s.nextBikeID = state.NextBikeID
	}
	if state.NextCustomerID > s.nextCustomerID {
		s.nextCustomerID = state.NextCustomerID
	}
	if state.NextBookingID > s.nextBookingID {
		s.nextBookingID = state.NextBookingID
	}
	s.hourlyRateCents = state.HourlyRateCents
	return nil
}

func (s *Store) CreateBike(bikeType string) (domain.Bike, error) {
	bikeType = strings.TrimSpace(bikeType)
	if bikeType == "" {
		return domain.Bike{}, fmt.Errorf("%w: Fahrradtyp darf nicht leer sein", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	bike := domain.Bike{ID: s.nextBikeID, Type: bikeType}
	s.bikes[bike.ID] = bike
	s.nextBikeID++
	if err := s.persistLocked(); err != nil {
		delete(s.bikes, bike.ID)
		s.nextBikeID--
		return domain.Bike{}, err
	}
	return bike, nil
}

func (s *Store) UpdateBike(id int, bikeType string) (domain.Bike, error) {
	bikeType = strings.TrimSpace(bikeType)
	if id < 1 || bikeType == "" {
		return domain.Bike{}, fmt.Errorf("%w: ID und Fahrradtyp müssen gültig sein", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	bike, ok := s.bikes[id]
	if !ok {
		return domain.Bike{}, ErrNotFound
	}
	previous := bike
	bike.Type = bikeType
	s.bikes[id] = bike
	if err := s.persistLocked(); err != nil {
		s.bikes[id] = previous
		return domain.Bike{}, err
	}
	return bike, nil
}

func (s *Store) ListBikes() []domain.Bike {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bikes := make([]domain.Bike, 0, len(s.bikes))
	for _, bike := range s.bikes {
		bikes = append(bikes, bike)
	}
	sort.Slice(bikes, func(i, j int) bool { return bikes[i].ID < bikes[j].ID })
	return bikes
}

func (s *Store) CreateCustomer(name, documentType, documentNumber string) (domain.Customer, error) {
	name, documentType, documentNumber = strings.TrimSpace(name), strings.TrimSpace(documentType), strings.TrimSpace(documentNumber)
	if name == "" || documentType == "" || documentNumber == "" {
		return domain.Customer{}, fmt.Errorf("%w: Name, Ausweistyp und Ausweisnummer sind erforderlich", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	customer := domain.Customer{
		ID:             s.nextCustomerID,
		Name:           name,
		DocumentType:   documentType,
		DocumentNumber: documentNumber,
	}
	s.customers[customer.ID] = customer
	s.nextCustomerID++
	if err := s.persistLocked(); err != nil {
		delete(s.customers, customer.ID)
		s.nextCustomerID--
		return domain.Customer{}, err
	}
	return customer, nil
}

func (s *Store) UpdateCustomer(id int, name, documentType, documentNumber string) (domain.Customer, error) {
	name, documentType, documentNumber = strings.TrimSpace(name), strings.TrimSpace(documentType), strings.TrimSpace(documentNumber)
	if id < 1 || name == "" || documentType == "" || documentNumber == "" {
		return domain.Customer{}, fmt.Errorf("%w: Name, Ausweistyp und Ausweisnummer sind erforderlich", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	customer, ok := s.customers[id]
	if !ok {
		return domain.Customer{}, ErrNotFound
	}
	previous := customer
	customer.Name, customer.DocumentType, customer.DocumentNumber = name, documentType, documentNumber
	s.customers[id] = customer
	if err := s.persistLocked(); err != nil {
		s.customers[id] = previous
		return domain.Customer{}, err
	}
	return customer, nil
}

func (s *Store) ListCustomers() []domain.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	customers := make([]domain.Customer, 0, len(s.customers))
	for _, customer := range s.customers {
		customers = append(customers, customer)
	}
	sort.Slice(customers, func(i, j int) bool { return customers[i].ID < customers[j].ID })
	return customers
}

func (s *Store) CreateBooking(bikeID, customerID int, start, end, now time.Time) (domain.Booking, error) {
	if bikeID < 1 || customerID < 1 || !end.After(start) {
		return domain.Booking{}, fmt.Errorf("%w: Fahrrad, Kunde und ein gültiger Mietzeitraum sind erforderlich", ErrInvalid)
	}
	if start.Before(now) {
		return domain.Booking{}, fmt.Errorf("%w: Buchungsbeginn darf nicht in der Vergangenheit liegen", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bikes[bikeID]; !ok {
		return domain.Booking{}, fmt.Errorf("Fahrrad %d: %w", bikeID, ErrNotFound)
	}
	if _, ok := s.customers[customerID]; !ok {
		return domain.Booking{}, fmt.Errorf("Kunde %d: %w", customerID, ErrNotFound)
	}
	for _, existing := range s.bookings {
		if existing.BikeID == bikeID && overlaps(start, end, existing.StartTime, effectiveEnd(existing)) {
			return domain.Booking{}, ErrUnavailable
		}
	}
	hours := billedHours(end.Sub(start))
	if s.hourlyRateCents > 0 && hours > math.MaxInt64/s.hourlyRateCents {
		return domain.Booking{}, fmt.Errorf("%w: Gesamtpreis ist zu hoch", ErrInvalid)
	}
	booking := domain.Booking{
		ID:              s.nextBookingID,
		BikeID:          bikeID,
		CustomerID:      customerID,
		StartTime:       start,
		EndTime:         end,
		HourlyRateCents: s.hourlyRateCents,
		TotalPriceCents: hours * s.hourlyRateCents,
	}
	s.bookings[booking.ID] = booking
	s.nextBookingID++
	if err := s.persistLocked(); err != nil {
		delete(s.bookings, booking.ID)
		s.nextBookingID--
		return domain.Booking{}, err
	}
	return booking, nil
}

func billedHours(duration time.Duration) int64 {
	hours := int64(duration / time.Hour)
	if duration%time.Hour != 0 {
		hours++
	}
	return hours
}

func overlaps(start, end, otherStart, otherEnd time.Time) bool {
	return start.Before(otherEnd) && otherStart.Before(end)
}

func effectiveEnd(booking domain.Booking) time.Time {
	if booking.ReturnTime != nil && booking.ReturnTime.Before(booking.EndTime) {
		return *booking.ReturnTime
	}
	return booking.EndTime
}

func (s *Store) ReturnBooking(id int, returnedAt time.Time) (domain.Booking, error) {
	if id < 1 {
		return domain.Booking{}, fmt.Errorf("%w: Buchungs-ID muss positiv sein", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	booking, ok := s.bookings[id]
	if !ok {
		return domain.Booking{}, ErrNotFound
	}
	if booking.ReturnTime != nil || returnedAt.Before(booking.StartTime) {
		return domain.Booking{}, fmt.Errorf("%w: Buchung kann derzeit nicht zurückgegeben werden", ErrInvalid)
	}
	returnTime := returnedAt
	booking.ReturnTime = &returnTime
	s.bookings[id] = booking
	if err := s.persistLocked(); err != nil {
		booking.ReturnTime = nil
		s.bookings[id] = booking
		return domain.Booking{}, err
	}
	return booking, nil
}

func (s *Store) ListBookings() []domain.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bookings := make([]domain.Booking, 0, len(s.bookings))
	for _, booking := range s.bookings {
		bookings = append(bookings, booking)
	}
	sort.Slice(bookings, func(i, j int) bool { return bookings[i].ID < bookings[j].ID })
	return bookings
}

func (s *Store) Dashboard(now time.Time) domain.Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dashboard := domain.Dashboard{TotalBikes: len(s.bikes), Customers: len(s.customers)}
	for _, bike := range s.bikes {
		rented := false
		for _, booking := range s.bookings {
			if booking.BikeID == bike.ID && booking.ReturnTime == nil &&
				!booking.StartTime.After(now) && now.Before(booking.EndTime) {
				rented = true
				break
			}
		}
		if rented {
			dashboard.RentedBikes++
		} else {
			dashboard.AvailableBikes++
		}
	}
	for _, booking := range s.bookings {
		if booking.ReturnTime != nil {
			continue
		}
		switch {
		case booking.StartTime.After(now):
			dashboard.Reservations++
		case now.Before(booking.EndTime):
			dashboard.ActiveRentals++
		}
	}
	return dashboard
}

func (s *Store) HourlyRateCents() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hourlyRateCents
}

func (s *Store) SetHourlyRateCents(rate int64) error {
	if rate < 0 {
		return fmt.Errorf("%w: Stundentarif darf nicht negativ sein", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.hourlyRateCents
	s.hourlyRateCents = rate
	if err := s.persistLocked(); err != nil {
		s.hourlyRateCents = previous
		return err
	}
	return nil
}

func (s *Store) persistLocked() error {
	if s.dataFile == "" {
		return nil
	}
	state := persistedState{
		Bikes:           make([]domain.Bike, 0, len(s.bikes)),
		Customers:       make([]domain.Customer, 0, len(s.customers)),
		Bookings:        make([]domain.Booking, 0, len(s.bookings)),
		HourlyRateCents: s.hourlyRateCents,
		NextBikeID:      s.nextBikeID,
		NextCustomerID:  s.nextCustomerID,
		NextBookingID:   s.nextBookingID,
	}
	for _, bike := range s.bikes {
		state.Bikes = append(state.Bikes, bike)
	}
	for _, customer := range s.customers {
		state.Customers = append(state.Customers, customer)
	}
	for _, booking := range s.bookings {
		state.Bookings = append(state.Bookings, booking)
	}
	sort.Slice(state.Bikes, func(i, j int) bool { return state.Bikes[i].ID < state.Bikes[j].ID })
	sort.Slice(state.Customers, func(i, j int) bool { return state.Customers[i].ID < state.Customers[j].ID })
	sort.Slice(state.Bookings, func(i, j int) bool { return state.Bookings[i].ID < state.Bookings[j].ID })

	directory := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("Datenverzeichnis konnte nicht erstellt werden: %w", err)
	}
	tempFile, err := os.CreateTemp(directory, ".fahrrad-verleih-*.tmp")
	if err != nil {
		return fmt.Errorf("temporäre Datendatei konnte nicht erstellt werden: %w", err)
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)
	if err := json.NewEncoder(tempFile).Encode(state); err != nil {
		tempFile.Close()
		return fmt.Errorf("Daten konnten nicht serialisiert werden: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("temporäre Datendatei konnte nicht synchronisiert werden: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("temporäre Datendatei konnte nicht geschlossen werden: %w", err)
	}
	if err := os.Rename(tempName, s.dataFile); err != nil {
		return fmt.Errorf("Datendatei konnte nicht ersetzt werden: %w", err)
	}
	return nil
}
