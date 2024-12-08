package main

import (
	"database/sql"
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

var users []User

func main() {
	fmt.Printf("Welcome to Gowtham App \n")
	r := mux.NewRouter()
	r.HandleFunc("/", welcomePage).Methods("GET")
	r.HandleFunc("/guest", CreateGifts).Methods("POST")
	r.HandleFunc("/create-tables", TableCreation).Methods("POST")
	r.HandleFunc("/guest/{name}", FilterUser).Methods("GET")
	r.HandleFunc("/native/{native}", FilterNative).Methods("GET")
	r.HandleFunc("/total-amount", GetTotalAmount).Methods("GET")
	r.HandleFunc("/delete-guest/{name}", DeleteUser).Methods("DELETE")
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

func TableCreation(w http.ResponseWriter, r *http.Request) {
	db, err := database.DatabaseConnection()
	if err != nil {
		log.Fatal(err)
		return
	}
	//Create table gifts
	query1 := `CREATE TABLE gifts.gifts (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		native VARCHAR(100) NOT NULL,
		amount INT NOT NULL,
		event_date VARCHAR(100),
		notes VARCHAR(100)
		);`

	_, err = db.Exec(query1)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error Creating Table gifts")
		return
	}

	//Create table backup
	query2 := `CREATE TABLE gifts.backup (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		native VARCHAR(100) NOT NULL,
		amount INT NOT NULL,
		event_date VARCHAR(100),
		notes VARCHAR(100)
		);`

	_, err = db.Exec(query2)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error Creating Table backup")
	}

	fmt.Println("The tables created successfully...")
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

	query1 := `INSERT INTO gifts (name, native, amount, event_date, notes) 
	          VALUES (?, ?, ?, NOW(), ?)`

	insert1, err := db.Prepare(query1)
	if err != nil {
		http.Error(w, "Error preparing the SQL statement", http.StatusInternalServerError)
		log.Println("Error preparing the SQL statement:", err)
		return
	}
	defer insert1.Close()

	// Use Exec to insert data into the table, without the ID since it's auto-generated.
	result1, err := insert1.Exec(user.Name, user.Native, user.Amount, user.Notes)
	if err != nil {
		http.Error(w, "Error inserting data into the database", http.StatusInternalServerError)
		log.Println("Error inserting data into the database:", err)
		return
	}

	// Get the ID of the newly inserted user
	userID, err := result1.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
		log.Println("Error retrieving the user ID:", err)
		return
	}

	//Updating data for backup
	query2 := `INSERT INTO backup (name, native, amount, event_date, notes) 
	          VALUES (?, ?, ?, NOW(), ?)`

	insert2, err := db.Prepare(query2)
	if err != nil {
		http.Error(w, "Error preparing the SQL statement", http.StatusInternalServerError)
		log.Println("Error preparing the SQL statement:", err)
		return
	}
	defer insert2.Close()

	// Use Exec to insert data into the table, without the ID since it's auto-generated.
	_, err = insert2.Exec(user.Name, user.Native, user.Amount, user.Notes)
	if err != nil {
		http.Error(w, "Error inserting data into the database for backup", http.StatusInternalServerError)
		log.Println("Error inserting data into the database forbackup:", err)
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

func FilterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	name := params["name"]

	// Establish database connection
	db, err := database.DatabaseConnection()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database for the user with the matching name
	query := `SELECT id, name, native, amount, event_date, notes FROM gifts.gifts WHERE name = ?`
	var user User

	// Execute the query with the name from the URL parameter
	err = db.QueryRow(query, name).Scan(&user.ID, &user.Name, &user.Native, &user.Amount, &user.EventDate, &user.Notes)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		log.Println("Error querying the database:", err)
		return
	}

	// Return the found user as a JSON response
	json.NewEncoder(w).Encode(user)
	fmt.Printf("Guest %v gifted the amount: %v\n", user.Name, user.Amount)
}

func FilterNative(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	native := params["native"]

	// Establish database connection
	db, err := database.DatabaseConnection()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database for the user with the matching native
	query := `SELECT id, name, native, amount, event_date, notes FROM gifts.gifts WHERE native = ?`
	//var user User

	// Execute the query with the native from the URL parameter
	rows, err := db.Query(query, native)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		log.Println("Error querying the database:", err)
		return
	}
	defer rows.Close()

	// Create a slice to hold all matching users
	var users []User

	// Loop through the rows and scan the data into the users slice
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Native, &user.Amount, &user.EventDate, &user.Notes)
		if err != nil {
			http.Error(w, "Error scanning the rows", http.StatusInternalServerError)
			log.Println("Error scanning the rows:", err)
			return
		}
		users = append(users, user)
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		http.Error(w, "Error processing the rows", http.StatusInternalServerError)
		log.Println("Error processing the rows:", err)
		return
	}

	// If no users are found, return a 404
	if len(users) == 0 {
		http.Error(w, "No users found with the given native", http.StatusNotFound)
		return
	}

	// Return the found user as a JSON response
	json.NewEncoder(w).Encode(users)
	fmt.Printf("List of users %v from %v\n\n", native, users)
}

func GetTotalAmount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Establish database connection
	db, err := database.DatabaseConnection()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `SELECT SUM(amount) AS total_amount FROM gifts.gifts`

	// Execute the query
	var totalAmount float64
	err = db.QueryRow(query).Scan(&totalAmount)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No data found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	// Return the sum as a JSON response
	response := map[string]float64{
		"total_amount": totalAmount,
	}

	json.NewEncoder(w).Encode(response)
	fmt.Printf("The total amount: %v", response)

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Establish database connection
	db, err := database.DatabaseConnection()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	query := `DELETE FROM gifts.gifts WHERE name = ?`
	params := mux.Vars(r)
	name := params["name"]
	result, err := db.Exec(query, name)
	if err != nil {
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		log.Println("Error querying the database:", err)
		return
	}
	// Check if any row was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking affected rows", http.StatusInternalServerError)
		log.Println("Error checking affected rows:", err)
		return
	}

	// If no rows were affected, it means the name wasn't found
	if rowsAffected == 0 {
		http.Error(w, "No matching gift found", http.StatusNotFound)
		return
	}

	// Return a success message with status 200 OK
	response := map[string]string{"message": fmt.Sprintf("Successfully deleted gift with name: %s", name)}
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
	successMessage := fmt.Sprintf("The gift from '%v' was deleted successfully.", name)
	err = notifyslack.sendToSlack(successMessage)
	if err != nil {
		log.Println("Error sending message to Slack:", err)
	}

}
