package main

import (
	"errors"
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

func generateCalendar(plan []mealPlanItem) (*ics.Calendar, error) {
	mealTimes := map[string]time.Duration{
		"breakfast": time.Hour * 9,
		"lunch":     time.Hour * 12,
		"dinner":    time.Hour * 19,
	}

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
			log.Fatalf("failed to parse date %v: %v", item.Date, err)
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

	cal, err := generateCalendar(mealPlan)
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

	http.HandleFunc("/mealie.ics", getCalendar)

	log.Println("listening on port 3333")
	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else if err != nil {
		log.Fatalf("web server failed: %v", err)
	}
}
