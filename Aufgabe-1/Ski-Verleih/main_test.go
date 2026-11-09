package main

import (
	"testing"

	"ski-verleih/internal/store"
)

func TestConfiguredDailyRateCents(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{name: "default rate", want: store.DefaultDailyRateCents},
		{name: "whole euros", value: "20", want: 2_000},
		{name: "one decimal place", value: "20.5", want: 2_050},
		{name: "two decimal places", value: "0.09", want: 9},
		{name: "empty value", value: "", wantErr: true},
		{name: "negative value", value: "-1", wantErr: true},
		{name: "signed value", value: "+1", wantErr: true},
		{name: "comma decimal", value: "20,50", wantErr: true},
		{name: "too many decimals", value: "20.001", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := parseDailyRateCents(testCase.value, testCase.name != "default rate")
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("configuredDailyRateCents() = %d, erwartet Fehler", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != testCase.want {
				t.Fatalf("configuredDailyRateCents() = %d, erwartet %d", got, testCase.want)
			}
		})
	}
}
