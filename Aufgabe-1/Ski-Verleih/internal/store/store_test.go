package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestIDsAreSequentialAndValuesCanBeUpdated(t *testing.T) {
	dataStore := New()

	firstSki, err := dataStore.CreateSki("All-Mountain")
	if err != nil {
		t.Fatal(err)
	}
	secondSki, err := dataStore.CreateSki("Carving")
	if err != nil {
		t.Fatal(err)
	}
	if firstSki.ID != 1 || secondSki.ID != 2 {
		t.Fatalf("Ski-IDs = %d, %d; erwartet 1, 2", firstSki.ID, secondSki.ID)
	}

	updated, err := dataStore.UpdateSki(firstSki.ID, "Freeride")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Type != "Freeride" || updated.ID != firstSki.ID {
		t.Fatalf("unerwartetes aktualisiertes Ski-Objekt: %+v", updated)
	}

	firstCustomer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}
	secondCustomer, err := dataStore.CreateCustomer("Sam Beispiel", "Personalausweis", "A98765")
	if err != nil {
		t.Fatal(err)
	}
	if firstCustomer.ID != 1 || secondCustomer.ID != 2 {
		t.Fatalf("Kunden-IDs = %d, %d; erwartet 1, 2", firstCustomer.ID, secondCustomer.ID)
	}

	updatedCustomer, err := dataStore.UpdateCustomer(firstCustomer.ID, "Alex B.", "Reisepass", "X54321")
	if err != nil {
		t.Fatal(err)
	}
	if updatedCustomer.Name != "Alex B." || updatedCustomer.DocumentNumber != "X54321" {
		t.Fatalf("Kundendaten wurden nicht aktualisiert: %+v", updatedCustomer)
	}
}

func TestCheckoutRequiresValidAvailableSkiAndWholeDays(t *testing.T) {
	dataStore := New()
	ski, err := dataStore.CreateSki("Touring")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}

	for _, days := range []int{0, -1} {
		if _, err := dataStore.Checkout(ski.ID, customer.ID, days, time.Now()); !errors.Is(err, ErrInvalid) {
			t.Errorf("Checkout mit %d Tagen: Fehler = %v, erwartet ErrInvalid", days, err)
		}
	}

	start := time.Date(2026, 11, 5, 16, 20, 0, 0, time.FixedZone("CET", 3600))
	rental, err := dataStore.Checkout(ski.ID, customer.ID, 3, start)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rental.StartDate.Format("2006-01-02 15:04"), "2026-11-05 00:00"; got != want {
		t.Fatalf("StartDate = %q; erwartet %q", got, want)
	}
	if got, want := rental.DueDate.Format("2006-01-02 15:04"), "2026-11-08 00:00"; got != want {
		t.Fatalf("DueDate = %q; erwartet %q", got, want)
	}
	if _, err := dataStore.Checkout(ski.ID, customer.ID, 1, start); !errors.Is(err, ErrAlreadyRented) {
		t.Fatalf("zweite gleichzeitige Ausleihe: Fehler = %v, erwartet ErrAlreadyRented", err)
	}

	dashboard := dataStore.Dashboard()
	if dashboard.AvailableSkis != 0 || dashboard.RentedSkis != 1 || dashboard.TotalSkis != 1 || dashboard.Customers != 1 {
		t.Fatalf("Dashboard nach Ausleihe ist falsch: %+v", dashboard)
	}
}

func TestReturnChangesStatusAndRetainsHistory(t *testing.T) {
	dataStore := New()
	ski, err := dataStore.CreateSki("Carving")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dataStore.Checkout(ski.ID, customer.ID, 1, time.Now()); err != nil {
		t.Fatal(err)
	}

	returnedAt := time.Date(2026, 11, 6, 10, 30, 0, 0, time.Local)
	rental, err := dataStore.ReturnSki(ski.ID, returnedAt)
	if err != nil {
		t.Fatal(err)
	}
	if rental.ReturnDate == nil || !rental.ReturnDate.Equal(returnedAt) {
		t.Fatalf("Rückgabezeitpunkt wurde nicht gespeichert: %+v", rental)
	}
	if _, err := dataStore.ReturnSki(ski.ID, returnedAt); !errors.Is(err, ErrNotRented) {
		t.Fatalf("doppelte Rückgabe: Fehler = %v, erwartet ErrNotRented", err)
	}
	if len(dataStore.ListRentals()) != 1 {
		t.Fatalf("Verleihhistorie hat %d Einträge, erwartet 1", len(dataStore.ListRentals()))
	}
	dashboard := dataStore.Dashboard()
	if dashboard.AvailableSkis != 1 || dashboard.RentedSkis != 0 {
		t.Fatalf("Dashboard nach Rückgabe ist falsch: %+v", dashboard)
	}
}

func TestConcurrentCreationProducesUniqueSequentialIDs(t *testing.T) {
	const count = 64
	dataStore := New()
	var waitGroup sync.WaitGroup
	errorsFound := make(chan error, count)
	for i := 0; i < count; i++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			if _, err := dataStore.CreateSki(fmt.Sprintf("Ski-%d", index)); err != nil {
				errorsFound <- err
			}
		}(i)
	}
	waitGroup.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Error(err)
	}

	skis := dataStore.ListSkis()
	if len(skis) != count {
		t.Fatalf("Anzahl Ski = %d; erwartet %d", len(skis), count)
	}
	for index, ski := range skis {
		if ski.ID != index+1 {
			t.Fatalf("Ski-IDs nicht lückenlos sortiert: Index %d hat ID %d", index, ski.ID)
		}
	}
}

func TestRequiredFieldsAreValidated(t *testing.T) {
	dataStore := New()
	if _, err := dataStore.CreateSki(" "); !errors.Is(err, ErrInvalid) {
		t.Fatalf("leerer Ski-Typ: Fehler = %v, erwartet ErrInvalid", err)
	}
	if _, err := dataStore.CreateCustomer("Alex", "", "123"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("fehlender Ausweistyp: Fehler = %v, erwartet ErrInvalid", err)
	}
	if _, err := dataStore.UpdateSki(99, "Touring"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unbekannter Ski: Fehler = %v, erwartet ErrNotFound", err)
	}
}

func TestDailyRateIsConfigurableAndRentalPriceIsStored(t *testing.T) {
	dataStore, err := NewWithDailyRateCents(2_000)
	if err != nil {
		t.Fatal(err)
	}
	ski, err := dataStore.CreateSki("Touring")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}

	rental, err := dataStore.Checkout(ski.ID, customer.ID, 3, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if rental.DailyRateCents != 2_000 || rental.TotalPriceCents != 6_000 {
		t.Fatalf("Ausleihpreis = %d Cent bei %d Cent/Tag; erwartet 6000 Cent bei 2000 Cent/Tag", rental.TotalPriceCents, rental.DailyRateCents)
	}

	if err := dataStore.SetDailyRateCents(2_350); err != nil {
		t.Fatal(err)
	}
	if got := dataStore.DailyRateCents(); got != 2_350 {
		t.Fatalf("konfigurierter Tagestarif = %d Cent; erwartet 2350", got)
	}
	if got := dataStore.ListRentals()[0].TotalPriceCents; got != 6_000 {
		t.Fatalf("bestehender Ausleihpreis wurde nach Tarifänderung verändert: %d Cent", got)
	}

	secondSki, err := dataStore.CreateSki("Carving")
	if err != nil {
		t.Fatal(err)
	}
	secondRental, err := dataStore.Checkout(secondSki.ID, customer.ID, 2, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if secondRental.TotalPriceCents != 4_700 {
		t.Fatalf("neuer Ausleihpreis = %d Cent; erwartet 4700", secondRental.TotalPriceCents)
	}
	if err := dataStore.SetDailyRateCents(-1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("negativer Tagestarif: Fehler = %v; erwartet ErrInvalid", err)
	}
}

func TestReservationsRejectOverlapsAndAllowAdjacentDates(t *testing.T) {
	dataStore, err := NewWithDailyRateCents(2_000)
	if err != nil {
		t.Fatal(err)
	}
	ski, err := dataStore.CreateSki("Touring")
	if err != nil {
		t.Fatal(err)
	}
	otherSki, err := dataStore.CreateSki("Carving")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 5, 15, 30, 0, 0, time.Local)
	firstStart := time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)
	first, err := dataStore.Reserve(ski.ID, customer.ID, 3, firstStart, now)
	if err != nil {
		t.Fatal(err)
	}
	if first.TotalPriceCents != 6_000 || !first.DueDate.Equal(time.Date(2026, 11, 13, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("Reservierung hat falschen Zeitraum oder Preis: %+v", first)
	}

	overlaps := []struct {
		name  string
		start time.Time
		days  int
	}{
		{name: "beginnt im bestehenden Zeitraum", start: time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local), days: 1},
		{name: "umschließt bestehenden Zeitraum", start: time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local), days: 5},
		{name: "beginnt vor und endet im bestehenden Zeitraum", start: time.Date(2026, 11, 8, 0, 0, 0, 0, time.Local), days: 3},
	}
	for _, testCase := range overlaps {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := dataStore.Reserve(ski.ID, customer.ID, testCase.days, testCase.start, now); !errors.Is(err, ErrUnavailable) {
				t.Fatalf("überlappende Reservierung: Fehler = %v; erwartet ErrUnavailable", err)
			}
		})
	}

	adjacent, err := dataStore.Reserve(ski.ID, customer.ID, 2, first.DueDate, now)
	if err != nil {
		t.Fatalf("Reservierung ab dem Fälligkeitstag wurde abgewiesen: %v", err)
	}
	if !adjacent.StartDate.Equal(first.DueDate) {
		t.Fatalf("Folgereservierung beginnt %v; erwartet %v", adjacent.StartDate, first.DueDate)
	}
	if _, err := dataStore.Reserve(otherSki.ID, customer.ID, 2, firstStart, now); err != nil {
		t.Fatalf("gleichzeitige Reservierung eines anderen Skis wurde abgewiesen: %v", err)
	}
	if _, err := dataStore.Reserve(ski.ID, customer.ID, 1, now, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Reservierung mit heutigem Start: Fehler = %v; erwartet ErrInvalid", err)
	}
	if _, err := dataStore.Reserve(ski.ID, customer.ID, 0, firstStart, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Reservierung mit null Tagen: Fehler = %v; erwartet ErrInvalid", err)
	}
}

func TestImmediateCheckoutCannotOverlapFutureReservation(t *testing.T) {
	dataStore := New()
	ski, err := dataStore.CreateSki("Touring")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.Local)
	reservationStart := time.Date(2026, 11, 7, 0, 0, 0, 0, time.Local)
	if _, err := dataStore.Reserve(ski.ID, customer.ID, 2, reservationStart, now); err != nil {
		t.Fatal(err)
	}
	if _, err := dataStore.Checkout(ski.ID, customer.ID, 3, now); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Ausleihe mit Konflikt zur Reservierung: Fehler = %v; erwartet ErrUnavailable", err)
	}
	if _, err := dataStore.Checkout(ski.ID, customer.ID, 1, now); err != nil {
		t.Fatalf("Ausleihe, die vor einer Reservierung endet, wurde abgewiesen: %v", err)
	}
}

func TestFutureReservationAllowedAfterCurrentRentalDueDate(t *testing.T) {
	dataStore := New()
	ski, err := dataStore.CreateSki("Rossignol")
	if err != nil {
		t.Fatal(err)
	}
	firstCustomer, err := dataStore.CreateCustomer("Kundin Eins", "Reisepass", "R-1")
	if err != nil {
		t.Fatal(err)
	}
	secondCustomer, err := dataStore.CreateCustomer("Kunde Zwei", "Personalausweis", "P-2")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 11, 5, 10, 0, 0, 0, time.Local)
	currentRental, err := dataStore.Checkout(ski.ID, firstCustomer.ID, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	wantDueDate := time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)
	if !currentRental.DueDate.Equal(wantDueDate) {
		t.Fatalf("Ausleihe fällig am %v; erwartet %v", currentRental.DueDate, wantDueDate)
	}

	futureStart := time.Date(2026, 11, 7, 0, 0, 0, 0, time.Local)
	reservation, err := dataStore.Reserve(ski.ID, secondCustomer.ID, 2, futureStart, now)
	if err != nil {
		t.Fatalf("Zukünftige Reservierung nach aktuellem Fälligkeitsdatum wurde abgewiesen: %v", err)
	}
	if reservation.SkiID != ski.ID || reservation.CustomerID != secondCustomer.ID || !reservation.StartDate.Equal(futureStart) {
		t.Fatalf("Reservierung hat falsche Daten: %+v", reservation)
	}

	overlappingStart := time.Date(2026, 11, 8, 0, 0, 0, 0, time.Local)
	if _, err := dataStore.Reserve(ski.ID, firstCustomer.ID, 1, overlappingStart, now); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Überlappende Reservierung nach aktueller Ausleihe: Fehler = %v; erwartet ErrUnavailable", err)
	}
}

func TestPersistentStoreSavesAndLoadsAllData(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "nested", "data.json")
	dataStore, err := NewPersistent(dataFile, 2_000)
	if err != nil {
		t.Fatal(err)
	}

	firstSki, err := dataStore.CreateSki("Rossignol")
	if err != nil {
		t.Fatal(err)
	}
	secondSki, err := dataStore.CreateSki("Atomic")
	if err != nil {
		t.Fatal(err)
	}
	firstCustomer, err := dataStore.CreateCustomer("Alex Beispiel", "Reisepass", "X12345")
	if err != nil {
		t.Fatal(err)
	}
	secondCustomer, err := dataStore.CreateCustomer("Sam Beispiel", "Personalausweis", "A98765")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := dataStore.UpdateSki(firstSki.ID, "Rossignol Hero"); err != nil {
		t.Fatal(err)
	}
	if _, err := dataStore.UpdateCustomer(firstCustomer.ID, "Alex B.", "Reisepass", "X54321"); err != nil {
		t.Fatal(err)
	}

	now := time.Now().In(time.Local)
	if _, err := dataStore.Checkout(firstSki.ID, firstCustomer.ID, 1, now); err != nil {
		t.Fatal(err)
	}
	returnedRental, err := dataStore.ReturnSki(firstSki.ID, now.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	reservationStart := dateOnly(now).AddDate(0, 0, 5)
	reservation, err := dataStore.Reserve(secondSki.ID, secondCustomer.ID, 3, reservationStart, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := dataStore.SetDailyRateCents(2_350); err != nil {
		t.Fatal(err)
	}

	fileContent, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("JSON-Datendatei wurde nicht geschrieben: %v", err)
	}
	var persisted persistedState
	if err := json.Unmarshal(fileContent, &persisted); err != nil {
		t.Fatalf("Datendatei enthält ungültiges JSON: %v", err)
	}
	if len(persisted.Skis) != 2 || len(persisted.Customers) != 2 || len(persisted.Rentals) != 2 {
		t.Fatalf("gespeicherte Datensätze unvollständig: Ski=%d Kunden=%d Ausleihen=%d", len(persisted.Skis), len(persisted.Customers), len(persisted.Rentals))
	}

	restored, err := NewPersistent(dataFile, DefaultDailyRateCents)
	if err != nil {
		t.Fatal(err)
	}
	if restored.DailyRateCents() != 2_350 {
		t.Fatalf("Tagestarif nach Laden = %d Cent; erwartet 2350", restored.DailyRateCents())
	}
	skis := restored.ListSkis()
	if len(skis) != 2 || skis[0].Type != "Rossignol Hero" || skis[1].Type != "Atomic" {
		t.Fatalf("Ski nach Laden unerwartet: %+v", skis)
	}
	customers := restored.ListCustomers()
	if len(customers) != 2 || customers[0].Name != "Alex B." || customers[0].DocumentNumber != "X54321" {
		t.Fatalf("Kunden nach Laden unerwartet: %+v", customers)
	}
	rentals := restored.ListRentals()
	if len(rentals) != 2 || rentals[0].ReturnDate == nil || !rentals[0].ReturnDate.Equal(*returnedRental.ReturnDate) {
		t.Fatalf("Rückgabehistorie nach Laden unerwartet: %+v", rentals)
	}
	if !rentals[1].StartDate.Equal(reservation.StartDate) || rentals[1].TotalPriceCents != reservation.TotalPriceCents {
		t.Fatalf("Reservierung oder Preis nach Laden unerwartet: %+v", rentals[1])
	}

	nextSki, err := restored.CreateSki("Fischer")
	if err != nil {
		t.Fatal(err)
	}
	nextCustomer, err := restored.CreateCustomer("Taylor Test", "Reisepass", "T12345")
	if err != nil {
		t.Fatal(err)
	}
	if nextSki.ID != 3 || nextCustomer.ID != 3 {
		t.Fatalf("IDs nach Neustart wurden nicht fortgesetzt: Ski=%d Kunde=%d", nextSki.ID, nextCustomer.ID)
	}
	nextRental, err := restored.Checkout(secondSki.ID, nextCustomer.ID, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if nextRental.ID != 3 {
		t.Fatalf("Ausleih-ID nach Neustart wurde nicht fortgesetzt: %d", nextRental.ID)
	}
}

func TestPersistentStoreRejectsInvalidJSON(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(dataFile, []byte(`{"skis":[`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPersistent(dataFile, DefaultDailyRateCents); err == nil {
		t.Fatal("NewPersistent() akzeptierte eine beschädigte JSON-Datei")
	}
}

func TestPersistentWriteFailureDoesNotApplyMutation(t *testing.T) {
	directory := t.TempDir()
	dataFile := filepath.Join(directory, "data.json")
	dataStore, err := NewPersistent(dataFile, DefaultDailyRateCents)
	if err != nil {
		t.Fatal(err)
	}

	blockedPath := filepath.Join(directory, "blocked")
	if err := os.Mkdir(blockedPath, 0755); err != nil {
		t.Fatal(err)
	}
	dataStore.dataFile = blockedPath
	if _, err := dataStore.CreateSki("Should not be saved"); err == nil {
		t.Fatal("CreateSki() meldete trotz Schreibfehler Erfolg")
	}
	if skis := dataStore.ListSkis(); len(skis) != 0 {
		t.Fatalf("nicht gespeicherte Mutation blieb im Store: %+v", skis)
	}

	dataStore.dataFile = dataFile
	ski, err := dataStore.CreateSki("Persisted")
	if err != nil {
		t.Fatal(err)
	}
	if ski.ID != 1 {
		t.Fatalf("fehlgeschlagene Speicherung hat ID verbraucht: %d", ski.ID)
	}
}
