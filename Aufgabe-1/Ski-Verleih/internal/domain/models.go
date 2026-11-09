package domain

import "time"

type Ski struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Rented bool   `json:"rented"`
}

type Customer struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
}

type Rental struct {
	ID              int        `json:"id"`
	SkiID           int        `json:"skiId"`
	CustomerID      int        `json:"customerId"`
	Days            int        `json:"days"`
	StartDate       time.Time  `json:"startDate"`
	DueDate         time.Time  `json:"dueDate"`
	DailyRateCents  int64      `json:"dailyRateCents"`
	TotalPriceCents int64      `json:"totalPriceCents"`
	ReturnDate      *time.Time `json:"returnDate,omitempty"`
}

type Dashboard struct {
	AvailableSkis int `json:"availableSkis"`
	RentedSkis    int `json:"rentedSkis"`
	TotalSkis     int `json:"totalSkis"`
	Customers     int `json:"customers"`
}
