package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

const (
	pageURL     = "https://receive-smss.com/"
	cookieName  = "__cfduid"
	smsEndpoint = "sms/"
)

// ScrapeAvailableNumbers Extracts the list of phone-numbers from the page
func ScrapeAvailableNumbers() ([]Number, error) {
	response, err := soup.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to make HTTP request to %s", pageURL)
	}
	numbers := make([]Number, 0)
	// scrape the page
	document := soup.HTMLParse(response)
	numbersContainer := document.Find("div", "class", "number-boxes")
	numberBoxes := numbersContainer.FindAll("div", "class", "number-boxes-item")
	for _, numberBox := range numberBoxes {
		numberElement := numberBox.FindStrict("div", "class", "row")
		if numberElement.Error == nil {
			numberContainer := numberElement.FindStrict("h4")
			countryContainer := numberElement.FindStrict("h5")
			if numberContainer.Error == nil && countryContainer.Error == nil {
				number := Number{
					CreatedAt: time.Now().Format("2006-01-02 15:04:05 Monday"),
					Number:    numberContainer.Text(),
					Country:   countryContainer.Text(),
				}
				numbers = append(numbers, number)
			}
		}
	}
	return numbers, nil
}

// ScrapeMessagesForNumber GET SMS from number
func ScrapeMessagesForNumber(number string) ([]Message, error) {
	// Get cookie first
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to make GET request: %v", err)
	}
	cookies := resp.Cookies()
	cookieValue := ""
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			cookieValue = cookie.Value
		}
	}
	// now use that value to set the cookie in soup
	soup.Cookie(cookieName, cookieValue)
	requestURL := pageURL + smsEndpoint + strings.ReplaceAll(number, "+", "") + "/"
	// make GET with soup:
	response, err := soup.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch data: %v", err)
	}
	document := soup.HTMLParse(response)
	table := document.Find("table")
	if table.Error != nil {
		return nil, fmt.Errorf("Failed to load messages: %v", table.Error)
	}
	tbody := table.Find("tbody")
	if tbody.Error != nil {
		return nil, fmt.Errorf("Failed to load messages: %v", tbody.Error)
	}
	tableRows := tbody.FindAll("tr")
	messages := make([]Message, 0)
	for _, row := range tableRows {
		cols := row.FindAll("td")
		if len(cols) < 3 {
			continue
		}
		message := Message{
			Originator: cols[0].FullText(),
			Body:       cols[1].FullText(),
			CreatedAt:  cols[2].FullText(),
		}
		messages = append(messages, message)
	}
	return messages, nil
}
