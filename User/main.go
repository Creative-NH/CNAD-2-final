package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

// Database connection function
func connectDB(dbUser, dbPassword, dbHost, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func main() {
	// Get database connection details from environment variables
	dbUser, dbPassword, dbHost, dbName := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME")
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" {
		log.Fatal("Database credentials not fully set in environment variables")
	}

	// Get local port from environment variables
	//localPort := os.Getenv("LOCAL_PORT")
	localPort := "5001"
	if localPort == "" {
		log.Fatal("LOCAL_PORT environment variable is not set")
	}

	// Connect to the database
	db, err := connectDB(dbUser, dbPassword, dbHost, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the router
	router := mux.NewRouter()

	// API Routes
	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		registrationHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/authenticate", func(w http.ResponseWriter, r *http.Request) {
		authenticationHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/getUserDetails", func(w http.ResponseWriter, r *http.Request) {
		getUserDetailsHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/updateUserDetails", func(w http.ResponseWriter, r *http.Request) {
		updateUserDetailsHandler(w, r, db)
	}).Methods("PUT")

	// CORS Configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Printf("Starting server on :%s", localPort)
	if err := http.ListenAndServe(":"+localPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type User struct {
	UserID      int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Email       string    `json:"email,omitempty"`
	Password    string    `json:"password,omitempty"`
	DateOfBirth time.Time `json:"dateOfBirth,omitempty"`
	PhoneNumber string    `json:"phoneNumber,omitempty"`
	Address     string    `json:"address,omitempty"`
}

// Custom JSON Unmarshaler for time.Time
func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User // Create an alias to avoid infinite recursion
	aux := &struct {
		DateOfBirth string `json:"dateOfBirth"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse date in YYYY-MM-DD format
	parsedDOB, err := time.Parse("2006-01-02", aux.DateOfBirth)
	if err != nil {
		return err
	}
	u.DateOfBirth = parsedDOB
	return nil
}

func validateUserInput(u User) map[string]string {
	errors := make(map[string]string)

	// Name validation
	if u.Name == "" {
		errors["name"] = "Name is required"
	}

	// Email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	reEmail := regexp.MustCompile(emailRegex)
	if u.Email == "" {
		errors["email"] = "Email is required"
	} else if !reEmail.MatchString(u.Email) {
		errors["email"] = "Invalid email format"
	}

	// Password validation
	if u.Password == "" || len(u.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters long"
	}

	// Date of Birth validation
	if u.DateOfBirth.Year() < 1900 || u.DateOfBirth.Year() > time.Now().Year() {
		log.Println("Invalid Date of Birth:", u.DateOfBirth)
		errors["dateOfBirth"] = "Invalid Date of Birth"
	}

	// Phone Number validation
	if u.PhoneNumber == "" {
		errors["phone"] = "Phone number is required"
	} else {
		phoneRegex := `^\+?[0-9]{7,15}$` // Allows optional '+' and 7-15 digits
		rePhone := regexp.MustCompile(phoneRegex)
		if u.PhoneNumber != "" && !rePhone.MatchString(u.PhoneNumber) {
			errors["phone"] = "Invalid phone number format"
		}
	}

	// Address validation
	if u.Address == "" {
		errors["address"] = "Address is required"
	}

	return errors
}

func registrationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Decode the incoming JSON request body
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		log.Println("JSON decoding error:", err)
		http.Error(w, `{"message":"Invalid input format"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input fields
	validationErrors := validateUserInput(u)

	// Check if email is already registered
	query := "SELECT UserID FROM Users WHERE Email = ?"
	var userID int
	err := db.QueryRow(query, u.Email).Scan(&userID)
	if err == sql.ErrNoRows {

	} else if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		validationErrors["email"] = "Email address already in use"
	}

	if len(validationErrors) > 0 {
		log.Println("Validation errors:", validationErrors)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hashing error:", err)
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Insert new user into the database
	query2 := "INSERT INTO Users(Name, Email, PasswordHash, DateOfBirth, PhoneNumber, Address) VALUES(?, ?, ?, ?, ?, ?)"
	// Insert User into DB
	_, err = db.Exec(query2, u.Name, u.Email, hashedPassword, u.DateOfBirth, u.PhoneNumber, u.Address)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Send success response
	response := map[string]string{
		"message": "Registration successful!.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func authenticationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Define a struct for authentication request
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the incoming JSON request body
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		log.Println("JSON decoding error:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate email and password
	if request.Email == "" || request.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Prepare SQL statement
	var storedUserID int
	var storedPassword string

	query := "SELECT UserID, PasswordHash FROM Users WHERE Email = ?"

	// Query the database for a user with the provided email
	err := db.QueryRow(query, request.Email).Scan(&storedUserID, &storedPassword)
	if err == sql.ErrNoRows {
		// If no user is found
		http.Error(w, "Invalid email", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Compare entered password with stored password
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(request.Password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Respond with success message
	response := map[string]interface{}{
		"message": "Login successful",
		"user_id": storedUserID, // Return User ID
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("JSON encoding error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getUserDetailsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Decode the incoming JSON request body
	var request struct {
		UserID int `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		log.Println("JSON decoding error:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate User ID
	if request.UserID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Query user details from the database
	var user struct {
		Name        string    `json:"name"`
		Email       string    `json:"email"`
		DateOfBirth time.Time `json:"date_of_birth"`
		PhoneNumber string    `json:"phone_number"`
		Address     string    `json:"address"`
	}

	query := "SELECT Name, Email, DateOfBirth, PhoneNumber, Address FROM Users WHERE UserID = ?"
	err := db.QueryRow(query, request.UserID).Scan(&user.Name, &user.Email, &user.DateOfBirth, &user.PhoneNumber, &user.Address)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Set secure response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Send response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Println("JSON encoding error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// receives user id and new email, dob, phone no. or address and updates record. returns update status
func updateUserDetailsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Decode the incoming JSON request body
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		log.Println("JSON decoding error:", err)
		http.Error(w, `{"message":"Invalid input"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input
	if u.UserID <= 0 {
		log.Println("Error: Invalid user ID received")
		http.Error(w, `{"message":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}
	u.Password = "Placeholder"
	validationErrors := validateUserInput(u)

	// Query the database for a user with the provided email
	query := "SELECT UserID FROM Users WHERE Email = ?"
	var storedUserID int
	err := db.QueryRow(query, u.Email).Scan(&storedUserID)
	if err == sql.ErrNoRows {

	} else if err != nil {
		log.Println("Database query error while checking email:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if storedUserID != u.UserID {
		validationErrors["email"] = "Email address already in use"
	}

	if len(validationErrors) > 0 {
		log.Println("Validation errors:", validationErrors)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Prepare the SQL statement
	query2 := "UPDATE Users SET Name=?, Email=?, DateOfBirth=?, PhoneNumber=?, Address=? WHERE UserID=?"
	result, err := db.Exec(query2, u.Name, u.Email, u.DateOfBirth, u.PhoneNumber, u.Address, u.UserID)
	if err != nil {
		log.Println("Database update error:", err)
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Check if any row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error checking rows affected:", err)
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No changes made or user with ID %d not found\n", u.UserID)
		http.Error(w, `{"message":"User not found or no changes made"}`, http.StatusNotFound)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}
