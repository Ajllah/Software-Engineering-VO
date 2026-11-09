package main

import (
	"testing"

	"fahrrad-verleih/internal/store"
)

func TestConfiguredHourlyRate(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		configured bool
		want       int64
		wantErr    bool
	}{
		{name: "default", want: store.DefaultHourlyRateCents},
		{name: "whole euros", value: "12", configured: true, want: 1200},
		{name: "one decimal", value: "5.5", configured: true, want: 550},
		{name: "two decimals", value: "0.05", configured: true, want: 5},
		{name: "negative", value: "-1", configured: true, wantErr: true},
		{name: "too many decimals", value: "1.001", configured: true, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseHourlyRateCents(test.value, test.configured)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseHourlyRateCents() error = %v, wantErr %v", err, test.wantErr)
			}
			if err == nil && got != test.want {
				t.Fatalf("parseHourlyRateCents() = %d, want %d", got, test.want)
			}
		})
	}
}
