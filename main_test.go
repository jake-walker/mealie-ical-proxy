package main

import (
	"strings"
	"testing"
	"time"
)

func TestLoadMealTimes(t *testing.T) {
	values := map[string]string{
		"MEALIE_BREAKFAST_TIME": "08:30",
		"MEALIE_DINNER_TIME":    "20:15",
	}

	times, err := loadMealTimes(func(key string) string {
		return values[key]
	})
	if err != nil {
		t.Fatalf("loadMealTimes() error = %v", err)
	}

	expected := map[string]time.Duration{
		"breakfast": 8*time.Hour + 30*time.Minute,
		"lunch":     12 * time.Hour,
		"dinner":    20*time.Hour + 15*time.Minute,
	}

	for meal, want := range expected {
		if got := times[meal]; got != want {
			t.Errorf("loadMealTimes()[%q] = %v, want %v", meal, got, want)
		}
	}
}

func TestLoadMealTimesRejectsInvalidValue(t *testing.T) {
	_, err := loadMealTimes(func(key string) string {
		if key == "MEALIE_LUNCH_TIME" {
			return "noon"
		}
		return ""
	})

	if err == nil {
		t.Fatal("loadMealTimes() error = nil, want invalid format error")
	}
}

func TestFormatMealTime(t *testing.T) {
	if got, want := formatMealTime(8*time.Hour+5*time.Minute), "08:05"; got != want {
		t.Errorf("formatMealTime() = %q, want %q", got, want)
	}
}

func TestGenerateCalendarUsesConfiguredMealTime(t *testing.T) {
	plan := []mealPlanItem{{
		Date:      "2026-06-08",
		EntryType: "breakfast",
		Id:        1,
		Title:     "Porridge",
	}}

	cal, err := generateCalendar(plan, map[string]time.Duration{
		"breakfast": 8*time.Hour + 30*time.Minute,
	})
	if err != nil {
		t.Fatalf("generateCalendar() error = %v", err)
	}

	if got := cal.Serialize(); !strings.Contains(got, "DTSTART:20260608T083000Z") {
		t.Errorf("calendar does not contain configured breakfast time:\n%s", got)
	}
}
