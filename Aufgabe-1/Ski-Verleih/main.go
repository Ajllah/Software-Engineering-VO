package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ski-verleih/internal/handler"
	"ski-verleih/internal/store"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	dailyRateCents, err := configuredDailyRateCents()
	if err != nil {
		log.Fatalf("Ungültiger Tagestarif: %v", err)
	}
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data.json"
	}
	dataStore, err := store.NewPersistent(dataFile, dailyRateCents)
	if err != nil {
		log.Fatalf("Store konnte nicht initialisiert werden: %v", err)
	}
	api := handler.NewAPI(dataStore)
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	webContent, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatalf("Frontend-Dateien konnten nicht geladen werden: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	address := os.Getenv("ADDR")
	if address == "" {
		address = "localhost:8080"
	}

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Ski-Verleih läuft unter http://%s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP-Server konnte nicht gestartet werden: %v", err)
	}
}

func configuredDailyRateCents() (int64, error) {
	value, configured := os.LookupEnv("DAILY_RATE_EUR")
	return parseDailyRateCents(value, configured)
}

func parseDailyRateCents(value string, configured bool) (int64, error) {
	if !configured {
		return store.DefaultDailyRateCents, nil
	}

	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("DAILY_RATE_EUR muss eine nichtnegative Dezimalzahl mit maximal zwei Nachkommastellen sein")
	}
	if len(parts) == 1 {
		parts = append(parts, "00")
	} else if len(parts[1]) > 2 {
		return 0, fmt.Errorf("DAILY_RATE_EUR darf höchstens zwei Nachkommastellen haben")
	} else if len(parts[1]) == 1 {
		parts[1] += "0"
	} else if len(parts[1]) == 0 {
		parts[1] = "00"
	}
	if !isDecimalDigits(parts[0]) || !isDecimalDigits(parts[1]) {
		return 0, fmt.Errorf("DAILY_RATE_EUR muss eine nichtnegative Dezimalzahl sein")
	}

	euros, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("DAILY_RATE_EUR ist ungültig oder zu groß")
	}
	cents, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("DAILY_RATE_EUR muss eine nichtnegative Dezimalzahl sein")
	}
	if euros > (int64(^uint64(0)>>1)-cents)/100 {
		return 0, fmt.Errorf("DAILY_RATE_EUR ist ungültig oder zu groß")
	}
	return euros*100 + cents, nil
}

func isDecimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
