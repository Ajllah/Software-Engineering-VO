package store

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func setupStore(t *testing.T, rate int64) (*Store, int, int) {
	t.Helper()
	s, err := NewWithHourlyRateCents(rate)
	if err != nil {
		t.Fatal(err)
	}
	bike, err := s.CreateBike("Citybike")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := s.CreateCustomer("Ada Beispiel", "Personalausweis", "A123")
	if err != nil {
		t.Fatal(err)
	}
	return s, bike.ID, customer.ID
}

func TestAutomaticIDsAndCustomerCanBookDifferentBikesAtSameTime(t *testing.T) {
	s, firstBikeID, customerID := setupStore(t, 500)
	secondBike, err := s.CreateBike("Trekkingrad")
	if err != nil {
		t.Fatal(err)
	}
	secondCustomer, err := s.CreateCustomer("Berta Beispiel", "Reisepass", "B456")
	if err != nil {
		t.Fatal(err)
	}
	if secondBike.ID != firstBikeID+1 || secondCustomer.ID != customerID+1 {
		t.Fatalf("IDs not sequential: bike=%d customer=%d", secondBike.ID, secondCustomer.ID)
	}
	start := time.Now().Add(time.Hour).Truncate(time.Minute)
	end := start.Add(2 * time.Hour)
	for _, bikeID := range []int{firstBikeID, secondBike.ID} {
		if _, err := s.CreateBooking(bikeID, customerID, start, end, time.Now()); err != nil {
			t.Fatalf("customer should be able to book bike %d: %v", bikeID, err)
		}
	}
}

func TestPriceRoundsUpEachStartedHourAndKeepsRate(t *testing.T) {
	s, bikeID, customerID := setupStore(t, 500)
	start := time.Now().Add(time.Hour).Truncate(time.Minute)
	booking, err := s.CreateBooking(bikeID, customerID, start, start.Add(70*time.Minute), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if booking.TotalPriceCents != 1000 || booking.HourlyRateCents != 500 {
		t.Fatalf("70 minutes at 500 cents/hour = %#v; want 1000 cents", booking)
	}
	if err := s.SetHourlyRateCents(700); err != nil {
		t.Fatal(err)
	}
	if got := s.ListBookings()[0].TotalPriceCents; got != 1000 {
		t.Fatalf("existing booking price changed to %d", got)
	}
}

func TestOverlapsRejectedButAdjacentIntervalsAllowed(t *testing.T) {
	s, bikeID, customerID := setupStore(t, 500)
	now := time.Now()
	start := now.Add(time.Hour).Truncate(time.Minute)
	end := start.Add(2 * time.Hour)
	if _, err := s.CreateBooking(bikeID, customerID, start, end, now); err != nil {
		t.Fatal(err)
	}
	_, err := s.CreateBooking(bikeID, customerID, start.Add(time.Hour), end.Add(time.Hour), now)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("overlapping booking error = %v, want ErrUnavailable", err)
	}
	if _, err := s.CreateBooking(bikeID, customerID, end, end.Add(time.Hour), now); err != nil {
		t.Fatalf("adjacent booking should be accepted: %v", err)
	}
}

func TestReturnMakesBikeAvailableFromActualReturnTime(t *testing.T) {
	s, bikeID, customerID := setupStore(t, 500)
	now := time.Now()
	start := now.Add(time.Hour).Truncate(time.Minute)
	end := start.Add(4 * time.Hour)
	booking, err := s.CreateBooking(bikeID, customerID, start, end, now)
	if err != nil {
		t.Fatal(err)
	}
	returnTime := start.Add(time.Hour)
	if _, err := s.ReturnBooking(booking.ID, returnTime); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateBooking(bikeID, customerID, returnTime, returnTime.Add(time.Hour), now); err != nil {
		t.Fatalf("booking should be possible from return time: %v", err)
	}
}

func TestInvalidDurationAndPastStartRejected(t *testing.T) {
	s, bikeID, customerID := setupStore(t, 500)
	now := time.Now()
	if _, err := s.CreateBooking(bikeID, customerID, now.Add(time.Hour), now.Add(time.Hour), now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("zero duration error = %v", err)
	}
	if _, err := s.CreateBooking(bikeID, customerID, now.Add(-time.Minute), now.Add(time.Hour), now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("past start error = %v", err)
	}
}

func TestConcurrentOverlappingBookingsOnlyOneSucceeds(t *testing.T) {
	s, bikeID, customerID := setupStore(t, 500)
	now := time.Now()
	start, end := now.Add(time.Hour), now.Add(2*time.Hour)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.CreateBooking(bikeID, customerID, start, end, now)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	successes, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrUnavailable):
			conflicts++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d, want one each", successes, conflicts)
	}
}

func TestPersistentStoreRestoresDataAndNextIDs(t *testing.T) {
	file := filepath.Join(t.TempDir(), "data.json")
	s, err := NewPersistent(file, 500)
	if err != nil {
		t.Fatal(err)
	}
	bike, err := s.CreateBike("E-Bike")
	if err != nil {
		t.Fatal(err)
	}
	customer, err := s.CreateCustomer("Ada Beispiel", "Personalausweis", "A123")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now().Add(time.Hour).Truncate(time.Minute)
	if _, err := s.CreateBooking(bike.ID, customer.ID, start, start.Add(70*time.Minute), time.Now()); err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewPersistent(file, 900)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.HourlyRateCents() != 500 || len(reloaded.ListBookings()) != 1 {
		t.Fatalf("state not restored: rate=%d bookings=%d", reloaded.HourlyRateCents(), len(reloaded.ListBookings()))
	}
	nextBike, err := reloaded.CreateBike("Lastenrad")
	if err != nil {
		t.Fatal(err)
	}
	if nextBike.ID != bike.ID+1 {
		t.Fatalf("next bike ID = %d, want %d", nextBike.ID, bike.ID+1)
	}
}
