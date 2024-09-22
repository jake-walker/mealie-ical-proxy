package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type mealPlanRecipe struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type mealPlanItem struct {
	Date      string          `json:"date"`
	EntryType string          `json:"entryType"`
	Id        int             `json:"id"`
	Recipe    *mealPlanRecipe `json:"recipe"`
	Title     string          `json:"title"`
}

type mealPlanResponse struct {
	Items      []mealPlanItem `json:"items"`
	Next       *string        `json:"next"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Previous   *string        `json:"previous"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

func getMealPlan(apiBase string, apiKey string) ([]mealPlanItem, error) {
	var items = []mealPlanItem{}

	page := 1

	for {
		u, err := url.JoinPath(apiBase, "/api/groups/mealplans")
		if err != nil {
			return items, err
		}

		u = fmt.Sprintf("%s?page=%d", u, page)

		log.Printf("fetching %v", u)

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return items, err
		}

		req.Header.Add("Authorization", fmt.Sprint("Bearer ", apiKey))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return items, err
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return items, err
		}

		resObject := mealPlanResponse{}
		err = json.Unmarshal(bytes, &resObject)
		if err != nil {
			return items, err
		}

		items = append(items, resObject.Items...)

		if page >= resObject.TotalPages {
			log.Printf("fetched %d items", len(items))
			return items, nil
		}

		page = page + 1
	}
}
