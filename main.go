package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Gowthamoptit/gift-backend/database"
	"github.com/gorilla/mux"
)

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Native    string `json:"native"`
	Amount    int    `json:"amount"`
	EventDate string `json:"event_date"`
	Notes     string `json:"notes"`
}

//var users []User

func main() {
	fmt.Printf("Welcome to Gowtham App \n")
	r := mux.NewRouter()
	r.HandleFunc("/", welcomePage).Methods("GET")
	r.HandleFunc("/user", CreateGifts).Methods("POST")
	fmt.Println("Server is starting at :4000")
	log.Fatal(http.ListenAndServe(":4000", r))

}

func welcomePage(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("welcome.html")
	if err != nil {
		// If the file is not found or there is another error, return an error response
		http.Error(w, "Could not read the HTML file", http.StatusInternalServerError)
		return
	}

	// Set the content type to "text/html" to tell the browser it's an HTML file
	w.Header().Set("Content-Type", "text/html")

	// Write the contents of the HTML file to the response
	w.Write(data)

}

func CreateGifts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User

	if r.Body == nil {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Update the User into Database
	db, err := database.DatabaseConnection()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	//defer db.Close() // Close DB after the function completes
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v\n", err)
		http.Error(w, "Error pinging the database", http.StatusInternalServerError)
		return
	}
	query := `INSERT INTO gifts (name, native, amount, event_date, notes) 
	          VALUES (?, ?, ?, NOW(), ?)`
	insert, err := db.Prepare(query)
	if err != nil {
		http.Error(w, "Error preparing the SQL statement", http.StatusInternalServerError)
		log.Println("Error preparing the SQL statement:", err)
		return
	}
	defer insert.Close()

	// Use Exec to insert data into the table, without the ID since it's auto-generated.
	result, err := insert.Exec(user.Name, user.Native, user.Amount, user.Notes)
	if err != nil {
		http.Error(w, "Error inserting data into the database", http.StatusInternalServerError)
		log.Println("Error inserting data into the database:", err)
		return
	}

	// Get the ID of the newly inserted user
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
		log.Println("Error retrieving the user ID:", err)
		return
	}

	// Add the user ID to the response
	userResponse := map[string]interface{}{
		"id":         userID,
		"name":       user.Name,
		"native":     user.Native,
		"amount":     user.Amount,
		"event_date": user.EventDate,
		"notes":      user.Notes,
	}

	// Respond with the created user and its ID in JSON format
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResponse)
	fmt.Printf("User created: %v\n", user.Name)
	defer db.Close()
}
