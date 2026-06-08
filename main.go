package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	ics "github.com/arran4/golang-ical"
)

var apiBase = ""
var apiKey = ""
var mealTimes = defaultMealTimes()

func defaultMealTimes() map[string]time.Duration {
	return map[string]time.Duration{
		"breakfast": time.Hour * 9,
		"lunch":     time.Hour * 12,
		"dinner":    time.Hour * 19,
	}
}

func loadMealTimes(getenv func(string) string) (map[string]time.Duration, error) {
	times := defaultMealTimes()
	envVars := map[string]string{
		"breakfast": "MEALIE_BREAKFAST_TIME",
		"lunch":     "MEALIE_LUNCH_TIME",
		"dinner":    "MEALIE_DINNER_TIME",
	}

	for meal, envVar := range envVars {
		value := getenv(envVar)
		if value == "" {
			continue
		}

		parsed, err := time.Parse("15:04", value)
		if err != nil {
			return nil, fmt.Errorf("%s must use HH:MM 24-hour format: %w", envVar, err)
		}

		times[meal] = time.Duration(parsed.Hour())*time.Hour + time.Duration(parsed.Minute())*time.Minute
	}

	return times, nil
}

func formatMealTime(mealTime time.Duration) string {
	hours := mealTime / time.Hour
	minutes := mealTime % time.Hour / time.Minute
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func generateCalendar(plan []mealPlanItem, mealTimes map[string]time.Duration) (*ics.Calendar, error) {

	cal := ics.NewCalendar()
	cal.SetName("Mealie Meal Plan")
	cal.SetProductId("-//Jake Walker//Mealie iCal Proxy")
	cal.SetMethod(ics.MethodRequest)

	for _, item := range plan {
		title := item.Title

		if item.Recipe != nil {
			title = item.Recipe.Name
		}

		date, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %v: %w", item.Date, err)
		}

		mealTime, ok := mealTimes[item.EntryType]
		if ok {
			date = date.Add(mealTime)
		}

		event := cal.AddEvent(strconv.Itoa(item.Id))
		event.SetDtStampTime(date)
		event.SetModifiedAt(time.Now())
		event.SetStartAt(date)
		event.SetEndAt(date.Add(1 * time.Hour))
		event.SetSummary(title)
	}

	return cal, nil
}

func getCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("got calendar request")

	mealPlan, err := getMealPlan(apiBase, apiKey)
	if err != nil {
		log.Fatalf("failed to fetch meal plan: %v", err)
	}

	cal, err := generateCalendar(mealPlan, mealTimes)
	if err != nil {
		log.Fatalf("failed to generate calendar: %v", err)
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"mealie.ics\"")
	io.WriteString(w, cal.Serialize())
}

func main() {
	apiBase = os.Getenv("MEALIE_URL")
	apiKey = os.Getenv("MEALIE_API_KEY")

	if apiBase == "" || apiKey == "" {
		log.Fatalln("the `MEALIE_URL` and `MEALIE_API_KEY` environment variables are required!")
	}

	var err error
	mealTimes, err = loadMealTimes(os.Getenv)
	if err != nil {
		log.Fatalf("invalid meal time configuration: %v", err)
	}

	log.Printf("configured Mealie URL: %s", apiBase)
	log.Printf(
		"configured meal times: breakfast=%s lunch=%s dinner=%s",
		formatMealTime(mealTimes["breakfast"]),
		formatMealTime(mealTimes["lunch"]),
		formatMealTime(mealTimes["dinner"]),
	)

	http.HandleFunc("/mealie.ics", getCalendar)

	log.Println("listening on port 3333")
	err = http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else if err != nil {
		log.Fatalf("web server failed: %v", err)
	}
}
