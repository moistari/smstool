package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/manifoldco/promptui"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ScrapeAvailableNumbers()
	for {
		idx, err := displayInitParameters()
		if err != nil {
			return err
		}
		switch idx {
		case 0:
			registerNumber()
			break
		case 1:
			listNumbers()
			break
		case 2:
			removeNumbers()
			break
		case 3:
			// check if filter needs to be enabled
			includeFilter, err := shouldIncludeFilter()
			if err != nil {
				return err
			}
			checkMessages(includeFilter)
			break
		case 4:
			fmt.Println("Bye!")
			return nil
		default:
			return fmt.Errorf("Option %d yet to be implemented: %v", idx, err)
		}
	}
}

// Number represents a new number to be addeded.
type Number struct {
	Country   string `json:"country"`
	Number    string `json:"number"`
	CreatedAt string `json:"created_at"`
}

// Message represents the message.
type Message struct {
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	Originator string `json:"originator"`
}

// Numbers are a list Number type.
type Numbers []Number

// Messages are a list of Message type.
type Messages []Message

// DB The database functions group
type DB struct{}

func (d *DB) getDBPath() (string, error) {
	// Look for the path to be specified in ENV FAKE_SMS_DB_DIR,
	// if not, use default $HOME as the path to create DB.
	// The DB will be created at <db_dir>/.fake-sms/db.json
	// If the DB does not exist, it will be created and will be
	// initialized to an empty array []
	dbPath, exists := os.LookupEnv("FAKE_SMS_DB_DIR")
	if !exists {
		dbPath = os.Getenv("HOME")
		dbPath = filepath.Join(dbPath, ".fake-sms")
	}
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dbPath, 0o700)
		if err != nil {
			return "", fmt.Errorf("Failed to create DB directory at %s: %v", dbPath, err)
		}
	}
	dbPath = filepath.Join(dbPath, "db.json")
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		emptyArray := []byte("[\n]\n")
		err = os.WriteFile(dbPath, emptyArray, 0o700)
		if err != nil {
			return "", fmt.Errorf("Faild to create DB file at %s: %v", dbPath, err)
		}
	}
	return dbPath, nil
}

func (d *DB) addToDB(number *Number) error {
	dbPath, err := d.getDBPath()
	if err != nil {
		return err
	}
	// read and serialize it to numbers
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("Failed to read DB file at %s: %v", dbPath, err)
	}
	// unmarshall the db to Numbers type
	numbers := Numbers{}
	err = json.Unmarshal(data, &numbers)
	if err != nil {
		return fmt.Errorf("Failed to de-serialize DB file %s: %v", dbPath, err)
	}
	numbers = append(numbers, *number)
	// write it back to the db
	data, err = json.Marshal(numbers)
	if err != nil {
		return fmt.Errorf("Failed to serialize DB file %s: %v", dbPath, err)
	}
	err = os.WriteFile(dbPath, data, 0o700)
	if err != nil {
		return fmt.Errorf("Failed to save DB file %s: %v", dbPath, err)
	}
	return nil
}

func (d *DB) getFromDB() (*Numbers, error) {
	dbPath, err := d.getDBPath()
	if err != nil {
		return nil, err
	}
	// read and serialize it to numbers
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read DB file at %s: %v", dbPath, err)
	}
	// unmarshall the db to Numbers type
	numbers := new(Numbers)
	err = json.Unmarshal(data, numbers)
	if err != nil {
		return nil, fmt.Errorf("Failed to de-serialize DB file %s: %v", dbPath, err)
	}
	return numbers, nil
}

func (d *DB) deleteFromDB(idx *int) error {
	dbPath, err := d.getDBPath()
	if err != nil {
		return err
	}
	// read and serialize it to numbers
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("Failed to read DB file at %s: %v", dbPath, err)
	}
	// unmarshall the db to Numbers type
	numbers := Numbers{}
	err = json.Unmarshal(data, &numbers)
	if err != nil {
		return fmt.Errorf("Failed to de-serialize DB file %s: %v", dbPath, err)
	}
	// delete by index
	if *idx > len(numbers)-1 {
		return errors.New("Number does not exist to be deleted from DB")
	}
	numbers = append(numbers[:*idx], numbers[*idx+1:]...)
	// serialize it back
	data, err = json.Marshal(numbers)
	if err != nil {
		return fmt.Errorf("Failed to serialize DB file %s: %v", dbPath, err)
	}
	err = os.WriteFile(dbPath, data, 0o700)
	if err != nil {
		return fmt.Errorf("Failed to save DB file %s: %v", dbPath, err)
	}
	return nil
}

func numbersToList(numbers *Numbers) *[]string {
	listOfNumbers := make([]string, len(*numbers))
	for idx, number := range *numbers {
		listOfNumbers[idx] = fmt.Sprintf("%s (%s)", number.Number, number.Country)
	}
	return &listOfNumbers
}

func displayInitParameters() (int, error) {
	prompt := promptui.Select{
		Label: "What you want to do?",
		Items: []string{"Add a new number", "List my numbers", "Remove a number", "Get my messages", "Exit"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return 0, err
	}
	// Return the index of the parameter selected
	return idx, nil
}

func getAvailNumbers() (*Numbers, error) {
	numArray, err := ScrapeAvailableNumbers()
	if err != nil {
		return nil, err
	}
	numbers := Numbers(numArray)
	return &numbers, nil
}

func registerNumber() error {
	numbers, err := getAvailNumbers()
	if err != nil {
		return err
	}
	if len(*numbers) == 0 {
		fmt.Println("No new numbers available right now")
	} else {
		numberList := numbersToList(numbers)
		// display numbers
		prompt := promptui.Select{
			Label: "These are the available numbers, choose any one of them",
			Items: *numberList,
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}
		if idx == -1 {
			fmt.Println("Nothing selected")
		} else {
			// new number selected, save it to the database file
			selectedNumber := &(*numbers)[idx]
			fmt.Printf("Selected %s, saving to database\n", selectedNumber)
			db := DB{}
			db.addToDB(selectedNumber)
		}
	}
	return nil
}

func listNumbers() error {
	db := DB{}
	numbers, err := db.getFromDB()
	if err != nil {
		return err
	}
	fmt.Println("Country\t\tNumber\t\tCreated At")
	fmt.Println("=======================================================================")
	for _, number := range *numbers {
		fmt.Printf("%s\t\t%s\t\t%s", number.Country, number.Number, number.CreatedAt)
	}
	return nil
}

func removeNumbers() error {
	db := DB{}
	numbers, err := db.getFromDB()
	if err != nil {
		return err
	}
	numberList := numbersToList(numbers)
	if len(*numberList) == 0 {
		return errors.New("No numbers saved to delete")
	}
	// display the list
	prompt := promptui.Select{
		Label: "These are the available numbers, choose any one of them",
		Items: *numberList,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return err
	}
	if idx == -1 {
		fmt.Println("Nothing selected")
	} else {
		// new number selected, save it to the database file
		selectedNumber := &(*numbers)[idx]
		fmt.Printf("Selected %s, removing from database\n", selectedNumber)
		db.deleteFromDB(&idx)
	}
	return nil
}

func messagePatternCheck(pattern *string, messages *Messages) (Messages, error) {
	r, err := regexp.Compile(*pattern)
	if err != nil {
		return nil, errors.New("Invalid regular expression provided")
	}
	filteredMessages := make([]Message, 0)
	for _, message := range *messages {
		// check match
		isMatch := r.Match([]byte(message.Body))
		if isMatch {
			filteredMessages = append(filteredMessages, message)
		}
	}
	return Messages(filteredMessages), nil
}

func checkMessages(enableFilter bool) error {
	var db DB
	numbers, err := db.getFromDB()
	if err != nil {
		return err
	}
	numberList := numbersToList(numbers)
	if len(*numberList) == 0 {
		return errors.New("No numbers saved to delete")
	}
	// display the list
	prompt := promptui.Select{
		Label: "These are the available numbers, choose any one of them",
		Items: *numberList,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return err
	}
	if idx == -1 {
		fmt.Println("Nothing selected")
	} else {
		// new number selected, save it to the database file
		selectedNumber := &(*numbers)[idx]
		fmt.Printf("Selected %s, fetching messages\n", selectedNumber)
		messagesArray, err := ScrapeMessagesForNumber(selectedNumber.Number)
		if err != nil {
			return err
		}
		// check message
		messages := Messages(messagesArray)
		// run filter if enabled:
		if enableFilter {
			fmt.Println("Enter the filter regular expression:")
			userFilterInput := ""
			fmt.Scanln(&userFilterInput)
			if userFilterInput == "" {
				userFilterInput = `.*`
			}
			// run the filter
			messages, err = messagePatternCheck(&userFilterInput, &messages)
			if err != nil {
				return err
			}
		}
		fmt.Println("===========================================")
		for _, message := range messages {
			fmt.Printf("Sender: %s, at: %s", message.Originator, message.CreatedAt)
			fmt.Printf("Body: %s", message.Body)
			fmt.Println("===========================================")
		}
		indentedData, _ := json.MarshalIndent(messages, "", "\t")
		// save the body as json
		fileName := fmt.Sprintf("%s.json", selectedNumber.Number)
		err = os.WriteFile(fileName, indentedData, 0o700)
		if err != nil {
			return fmt.Errorf("Failed to save file %s: %v", fileName, err)
		}
	}
	return nil
}

func shouldIncludeFilter() (bool, error) {
	prompt := promptui.Select{
		Label: "Do you want to filter the messages?",
		Items: []string{"Yes", "No"},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return false, errors.New("Failed to render prompt")
	}
	return idx == 0, nil
}
