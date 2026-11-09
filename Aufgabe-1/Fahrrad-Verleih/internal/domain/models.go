package domain

import "time"

type Bike struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type Customer struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
}

type Booking struct {
	ID              int        `json:"id"`
	BikeID          int        `json:"bikeId"`
	CustomerID      int        `json:"customerId"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         time.Time  `json:"endTime"`
	HourlyRateCents int64      `json:"hourlyRateCents"`
	TotalPriceCents int64      `json:"totalPriceCents"`
	ReturnTime      *time.Time `json:"returnTime,omitempty"`
}

type Dashboard struct {
	AvailableBikes int `json:"availableBikes"`
	RentedBikes    int `json:"rentedBikes"`
	TotalBikes     int `json:"totalBikes"`
	Customers      int `json:"customers"`
	Reservations   int `json:"reservations"`
	ActiveRentals  int `json:"activeRentals"`
}
