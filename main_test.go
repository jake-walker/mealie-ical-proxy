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

func TestLoadTimezoneDefaultsToUTC(t *testing.T) {
	location, err := loadTimezone("")
	if err != nil {
		t.Fatalf("loadTimezone() error = %v", err)
	}
	if location != time.UTC {
		t.Errorf("loadTimezone() = %v, want UTC", location)
	}
}

func TestLoadTimezoneRejectsInvalidValue(t *testing.T) {
	if _, err := loadTimezone("Not/A_Timezone"); err == nil {
		t.Fatal("loadTimezone() error = nil, want invalid timezone error")
	}
}

func TestLoadFloatingTimes(t *testing.T) {
	floating, err := loadFloatingTimes("true")
	if err != nil {
		t.Fatalf("loadFloatingTimes() error = %v", err)
	}
	if !floating {
		t.Error("loadFloatingTimes() = false, want true")
	}
}

func TestLoadFloatingTimesRejectsInvalidValue(t *testing.T) {
	if _, err := loadFloatingTimes("sometimes"); err == nil {
		t.Fatal("loadFloatingTimes() error = nil, want invalid boolean error")
	}
}

func TestGenerateCalendarUsesConfiguredMealTimeInUTCByDefault(t *testing.T) {
	plan := []mealPlanItem{{
		Date:      "2026-06-08",
		EntryType: "breakfast",
		Id:        1,
		Title:     "Porridge",
	}}

	cal, err := generateCalendar(plan, map[string]time.Duration{
		"breakfast": 8*time.Hour + 30*time.Minute,
	}, time.UTC, false)
	if err != nil {
		t.Fatalf("generateCalendar() error = %v", err)
	}

	if got := cal.Serialize(); !strings.Contains(got, "DTSTART:20260608T083000Z") {
		t.Errorf("calendar does not contain configured breakfast time:\n%s", got)
	}
}

func TestGenerateCalendarConvertsConfiguredTimezoneToUTC(t *testing.T) {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("time.LoadLocation() error = %v", err)
	}

	plan := []mealPlanItem{{
		Date:      "2026-06-08",
		EntryType: "lunch",
		Id:        1,
		Title:     "Lunch",
	}}

	cal, err := generateCalendar(plan, map[string]time.Duration{
		"lunch": 12 * time.Hour,
	}, location, false)
	if err != nil {
		t.Fatalf("generateCalendar() error = %v", err)
	}

	if got := cal.Serialize(); !strings.Contains(got, "DTSTART:20260608T100000Z") {
		t.Errorf("calendar does not contain timezone-adjusted lunch time:\n%s", got)
	}
}

func TestGenerateCalendarCanUseFloatingTimes(t *testing.T) {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("time.LoadLocation() error = %v", err)
	}

	plan := []mealPlanItem{{
		Date:      "2026-06-08",
		EntryType: "dinner",
		Id:        1,
		Title:     "Dinner",
	}}

	cal, err := generateCalendar(plan, map[string]time.Duration{
		"dinner": 19 * time.Hour,
	}, location, true)
	if err != nil {
		t.Fatalf("generateCalendar() error = %v", err)
	}

	serialized := cal.Serialize()
	if !strings.Contains(serialized, "DTSTART:20260608T190000\r\n") {
		t.Errorf("calendar does not contain floating dinner time:\n%s", serialized)
	}
	if strings.Contains(serialized, "DTSTART:20260608T190000Z") {
		t.Errorf("floating dinner time unexpectedly has UTC suffix:\n%s", serialized)
	}
}

func TestDateAtMealTimePreservesWallClockTimeAcrossDST(t *testing.T) {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("time.LoadLocation() error = %v", err)
	}

	got, err := dateAtMealTime("2026-03-29", 19*time.Hour, location)
	if err != nil {
		t.Fatalf("dateAtMealTime() error = %v", err)
	}

	if got.Hour() != 19 {
		t.Errorf("dateAtMealTime() hour = %d, want 19", got.Hour())
	}
}
