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

	"ski-verleih/internal/domain"
)

var (
	ErrNotFound      = errors.New("Eintrag nicht gefunden")
	ErrAlreadyRented = errors.New("Ski ist bereits ausgeliehen")
	ErrNotRented     = errors.New("Ski ist nicht ausgeliehen")
	ErrInvalid       = errors.New("ungültige Eingabe")
	ErrUnavailable   = errors.New("Ski ist im gewählten Zeitraum nicht verfügbar")
)

const DefaultDailyRateCents int64 = 2000

type Store struct {
	mu             sync.RWMutex
	skis           map[int]domain.Ski
	customers      map[int]domain.Customer
	rentals        map[int]domain.Rental
	dailyRateCents int64
	nextSkiID      int
	nextCustomerID int
	nextRentalID   int
	dataFile       string
}

type persistedState struct {
	Skis           []domain.Ski      `json:"skis"`
	Customers      []domain.Customer `json:"customers"`
	Rentals        []domain.Rental   `json:"rentals"`
	DailyRateCents int64             `json:"dailyRateCents"`
	NextSkiID      int               `json:"nextSkiId"`
	NextCustomerID int               `json:"nextCustomerId"`
	NextRentalID   int               `json:"nextRentalId"`
}

func New() *Store {
	return &Store{
		skis:           make(map[int]domain.Ski),
		customers:      make(map[int]domain.Customer),
		rentals:        make(map[int]domain.Rental),
		dailyRateCents: DefaultDailyRateCents,
		nextSkiID:      1,
		nextCustomerID: 1,
		nextRentalID:   1,
	}
}

func NewWithDailyRateCents(dailyRateCents int64) (*Store, error) {
	if dailyRateCents < 0 {
		return nil, fmt.Errorf("%w: Tagestarif darf nicht negativ sein", ErrInvalid)
	}
	return &Store{
		skis:           make(map[int]domain.Ski),
		customers:      make(map[int]domain.Customer),
		rentals:        make(map[int]domain.Rental),
		dailyRateCents: dailyRateCents,
		nextSkiID:      1,
		nextCustomerID: 1,
		nextRentalID:   1,
	}, nil
}

func NewPersistent(dataFile string, initialDailyRateCents int64) (*Store, error) {
	if strings.TrimSpace(dataFile) == "" {
		return nil, fmt.Errorf("JSON-Datenpfad darf nicht leer sein")
	}
	dataStore, err := NewWithDailyRateCents(initialDailyRateCents)
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
	if state.DailyRateCents < 0 {
		return fmt.Errorf("%w: negativer Tagestarif", ErrInvalid)
	}

	for _, ski := range state.Skis {
		if ski.ID < 1 || strings.TrimSpace(ski.Type) == "" {
			return fmt.Errorf("%w: ungültiger Ski in Datendatei", ErrInvalid)
		}
		if _, exists := s.skis[ski.ID]; exists {
			return fmt.Errorf("%w: Ski-ID %d mehrfach vorhanden", ErrInvalid, ski.ID)
		}
		ski.Rented = false
		s.skis[ski.ID] = ski
		if ski.ID >= s.nextSkiID {
			s.nextSkiID = ski.ID + 1
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
	for _, rental := range state.Rentals {
		if rental.ID < 1 || rental.SkiID < 1 || rental.CustomerID < 1 || rental.Days < 1 ||
			rental.StartDate.IsZero() || rental.DueDate.IsZero() ||
			rental.DueDate.Before(rental.StartDate) ||
			rental.DailyRateCents < 0 || rental.TotalPriceCents < 0 {
			return fmt.Errorf("%w: ungültige Ausleihe in Datendatei", ErrInvalid)
		}
		if _, exists := s.rentals[rental.ID]; exists {
			return fmt.Errorf("%w: Ausleih-ID %d mehrfach vorhanden", ErrInvalid, rental.ID)
		}
		if _, exists := s.skis[rental.SkiID]; !exists {
			return fmt.Errorf("%w: Ski %d für Ausleihe %d fehlt", ErrInvalid, rental.SkiID, rental.ID)
		}
		if _, exists := s.customers[rental.CustomerID]; !exists {
			return fmt.Errorf("%w: Kunde %d für Ausleihe %d fehlt", ErrInvalid, rental.CustomerID, rental.ID)
		}
		s.rentals[rental.ID] = rental
		if rental.ID >= s.nextRentalID {
			s.nextRentalID = rental.ID + 1
		}
	}

	if state.NextSkiID > s.nextSkiID {
		s.nextSkiID = state.NextSkiID
	}
	if state.NextCustomerID > s.nextCustomerID {
		s.nextCustomerID = state.NextCustomerID
	}
	if state.NextRentalID > s.nextRentalID {
		s.nextRentalID = state.NextRentalID
	}
	s.dailyRateCents = state.DailyRateCents
	return nil
}

func (s *Store) CreateSki(skiType string) (domain.Ski, error) {
	skiType = strings.TrimSpace(skiType)
	if skiType == "" {
		return domain.Ski{}, fmt.Errorf("%w: Ski-Typ darf nicht leer sein", ErrInvalid)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ski := domain.Ski{ID: s.nextSkiID, Type: skiType}
	s.skis[ski.ID] = ski
	s.nextSkiID++
	if err := s.persistLocked(); err != nil {
		delete(s.skis, ski.ID)
		s.nextSkiID--
		return domain.Ski{}, err
	}
	return ski, nil
}

func (s *Store) UpdateSki(id int, skiType string) (domain.Ski, error) {
	skiType = strings.TrimSpace(skiType)
	if id < 1 || skiType == "" {
		return domain.Ski{}, fmt.Errorf("%w: ID und Ski-Typ müssen gültig sein", ErrInvalid)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ski, ok := s.skis[id]
	if !ok {
		return domain.Ski{}, ErrNotFound
	}
	previousSki := ski
	ski.Type = skiType
	s.skis[id] = ski
	if err := s.persistLocked(); err != nil {
		s.skis[id] = previousSki
		return domain.Ski{}, err
	}
	return ski, nil
}

func (s *Store) ListSkis() []domain.Ski {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := dateOnly(time.Now())
	skis := make([]domain.Ski, 0, len(s.skis))
	for _, ski := range s.skis {
		ski.Rented = s.isRentedAtLocked(ski.ID, today)
		skis = append(skis, ski)
	}
	sort.Slice(skis, func(i, j int) bool { return skis[i].ID < skis[j].ID })
	return skis
}

func (s *Store) CreateCustomer(name, documentType, documentNumber string) (domain.Customer, error) {
	name = strings.TrimSpace(name)
	documentType = strings.TrimSpace(documentType)
	documentNumber = strings.TrimSpace(documentNumber)
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
	name = strings.TrimSpace(name)
	documentType = strings.TrimSpace(documentType)
	documentNumber = strings.TrimSpace(documentNumber)
	if id < 1 || name == "" || documentType == "" || documentNumber == "" {
		return domain.Customer{}, fmt.Errorf("%w: Name, Ausweistyp und Ausweisnummer sind erforderlich", ErrInvalid)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	customer, ok := s.customers[id]
	if !ok {
		return domain.Customer{}, ErrNotFound
	}
	previousCustomer := customer
	customer.Name = name
	customer.DocumentType = documentType
	customer.DocumentNumber = documentNumber
	s.customers[id] = customer
	if err := s.persistLocked(); err != nil {
		s.customers[id] = previousCustomer
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

func (s *Store) Checkout(skiID, customerID, days int, now time.Time) (domain.Rental, error) {
	return s.createRental(skiID, customerID, days, dateOnly(now), now, false)
}

func (s *Store) Reserve(skiID, customerID, days int, startDate, now time.Time) (domain.Rental, error) {
	startDate = dateOnly(startDate)
	today := dateOnly(now)
	if !startDate.After(today) {
		return domain.Rental{}, fmt.Errorf("%w: Reservierungsbeginn muss in der Zukunft liegen", ErrInvalid)
	}
	return s.createRental(skiID, customerID, days, startDate, now, true)
}

func (s *Store) createRental(skiID, customerID, days int, startDate, now time.Time, reservation bool) (domain.Rental, error) {
	if skiID < 1 || customerID < 1 || days < 1 {
		return domain.Rental{}, fmt.Errorf("%w: Ski, Kunde und positive ganze Tagesdauer sind erforderlich", ErrInvalid)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.skis[skiID]; !ok {
		return domain.Rental{}, fmt.Errorf("Ski %d: %w", skiID, ErrNotFound)
	}
	if _, ok := s.customers[customerID]; !ok {
		return domain.Rental{}, fmt.Errorf("Kunde %d: %w", customerID, ErrNotFound)
	}

	if int64(days) > 0 && s.dailyRateCents > 0 && int64(days) > math.MaxInt64/s.dailyRateCents {
		return domain.Rental{}, fmt.Errorf("%w: Gesamtpreis ist zu hoch", ErrInvalid)
	}
	dueDate := startDate.AddDate(0, 0, days)
	today := dateOnly(now)
	if s.hasOverlappingRentalLocked(skiID, startDate, dueDate) {
		if !reservation {
			if s.isRentedAtLocked(skiID, today) {
				return domain.Rental{}, ErrAlreadyRented
			}
		}
		return domain.Rental{}, ErrUnavailable
	}

	totalPriceCents := s.dailyRateCents * int64(days)
	rental := domain.Rental{
		ID:              s.nextRentalID,
		SkiID:           skiID,
		CustomerID:      customerID,
		Days:            days,
		StartDate:       startDate,
		DueDate:         dueDate,
		DailyRateCents:  s.dailyRateCents,
		TotalPriceCents: totalPriceCents,
	}
	s.rentals[rental.ID] = rental
	s.nextRentalID++
	if err := s.persistLocked(); err != nil {
		delete(s.rentals, rental.ID)
		s.nextRentalID--
		return domain.Rental{}, err
	}
	return rental, nil
}

func (s *Store) ReturnSki(skiID int, now time.Time) (domain.Rental, error) {
	if skiID < 1 {
		return domain.Rental{}, fmt.Errorf("%w: Ski-ID muss positiv sein", ErrInvalid)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.skis[skiID]
	if !ok {
		return domain.Rental{}, fmt.Errorf("Ski %d: %w", skiID, ErrNotFound)
	}

	var active domain.Rental
	found := false
	today := dateOnly(now)
	for _, rental := range s.rentals {
		if rental.SkiID == skiID && rental.ReturnDate == nil && !dateOnly(rental.StartDate).After(today) {
			active = rental
			found = true
			break
		}
	}
	if !found {
		return domain.Rental{}, ErrNotRented
	}

	returnDate := now
	active.ReturnDate = &returnDate
	s.rentals[active.ID] = active
	if err := s.persistLocked(); err != nil {
		active.ReturnDate = nil
		s.rentals[active.ID] = active
		return domain.Rental{}, err
	}
	return active, nil
}

func (s *Store) ListRentals() []domain.Rental {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rentals := make([]domain.Rental, 0, len(s.rentals))
	for _, rental := range s.rentals {
		rentals = append(rentals, rental)
	}
	sort.Slice(rentals, func(i, j int) bool { return rentals[i].ID < rentals[j].ID })
	return rentals
}

func (s *Store) Dashboard() domain.Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := dateOnly(time.Now())
	dashboard := domain.Dashboard{
		TotalSkis: len(s.skis),
		Customers: len(s.customers),
	}
	for _, ski := range s.skis {
		if s.isRentedAtLocked(ski.ID, today) {
			dashboard.RentedSkis++
		} else {
			dashboard.AvailableSkis++
		}
	}
	return dashboard
}

func (s *Store) DailyRateCents() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dailyRateCents
}

func (s *Store) SetDailyRateCents(dailyRateCents int64) error {
	if dailyRateCents < 0 {
		return fmt.Errorf("%w: Tagestarif darf nicht negativ sein", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	previousRate := s.dailyRateCents
	s.dailyRateCents = dailyRateCents
	if err := s.persistLocked(); err != nil {
		s.dailyRateCents = previousRate
		return err
	}
	return nil
}

func (s *Store) persistLocked() error {
	if s.dataFile == "" {
		return nil
	}

	state := persistedState{
		Skis:           make([]domain.Ski, 0, len(s.skis)),
		Customers:      make([]domain.Customer, 0, len(s.customers)),
		Rentals:        make([]domain.Rental, 0, len(s.rentals)),
		DailyRateCents: s.dailyRateCents,
		NextSkiID:      s.nextSkiID,
		NextCustomerID: s.nextCustomerID,
		NextRentalID:   s.nextRentalID,
	}
	for _, ski := range s.skis {
		state.Skis = append(state.Skis, ski)
	}
	for _, customer := range s.customers {
		state.Customers = append(state.Customers, customer)
	}
	for _, rental := range s.rentals {
		state.Rentals = append(state.Rentals, rental)
	}
	sort.Slice(state.Skis, func(i, j int) bool { return state.Skis[i].ID < state.Skis[j].ID })
	sort.Slice(state.Customers, func(i, j int) bool { return state.Customers[i].ID < state.Customers[j].ID })
	sort.Slice(state.Rentals, func(i, j int) bool { return state.Rentals[i].ID < state.Rentals[j].ID })

	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("Daten konnten nicht als JSON kodiert werden: %w", err)
	}
	directory := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("Datenverzeichnis %q konnte nicht angelegt werden: %w", directory, err)
	}
	temporaryFile, err := os.CreateTemp(directory, ".ski-verleih-*.tmp")
	if err != nil {
		return fmt.Errorf("temporäre Datendatei konnte nicht angelegt werden: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporaryFile.Write(content); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("temporäre Datendatei konnte nicht geschrieben werden: %w", err)
	}
	if err := temporaryFile.Sync(); err != nil {
		_ = temporaryFile.Close()
		return fmt.Errorf("temporäre Datendatei konnte nicht synchronisiert werden: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("temporäre Datendatei konnte nicht geschlossen werden: %w", err)
	}
	if err := os.Rename(temporaryPath, s.dataFile); err != nil {
		return fmt.Errorf("Datendatei %q konnte nicht aktualisiert werden: %w", s.dataFile, err)
	}
	return nil
}

func (s *Store) hasOverlappingRentalLocked(skiID int, startDate, dueDate time.Time) bool {
	for _, rental := range s.rentals {
		if rental.SkiID != skiID || rental.ReturnDate != nil {
			continue
		}

		existingStart := dateOnly(rental.StartDate)
		existingDue := dateOnly(rental.DueDate)
		if startDate.Before(existingDue) && existingStart.Before(dueDate) {
			return true
		}
	}
	return false
}

func (s *Store) isRentedAtLocked(skiID int, today time.Time) bool {
	for _, rental := range s.rentals {
		if rental.SkiID == skiID && rental.ReturnDate == nil && !dateOnly(rental.StartDate).After(today) {
			return true
		}
	}
	return false
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}
