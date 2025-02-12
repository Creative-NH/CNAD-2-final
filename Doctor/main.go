package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

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
	localPort := "5004"
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
	router.HandleFunc("/api/authenticate", func(w http.ResponseWriter, r *http.Request) {
		authenticationHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/getDoctorDetails", func(w http.ResponseWriter, r *http.Request) {
		getDoctorDetailsHandler(w, r, db)
	}).Methods("POST")

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

type Doctor struct {
	DoctorID int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func authenticationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse request body
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Println("Invalid JSON request:", err)
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate email format
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if match, _ := regexp.MatchString(emailRegex, credentials.Email); !match {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Query the database for the doctor with this email
	var doctor Doctor
	query := `SELECT DoctorID, PasswordHash FROM Doctors WHERE Email = ?`
	err := db.QueryRow(query, credentials.Email).Scan(&doctor.DoctorID, &doctor.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(doctor.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Authentication successful - Send doctor details (excluding password)
	response := map[string]interface{}{
		"message":   "Authentication successful",
		"doctor_id": doctor.DoctorID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getDoctorDetailsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse request body
	var request struct {
		DoctorID int `json:"doctor_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Println("Invalid JSON request:", err)
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate Doctor ID
	if request.DoctorID <= 0 {
		http.Error(w, "Invalid Doctor ID", http.StatusBadRequest)
		return
	}

	// Query doctor details from the database
	var doctor Doctor
	query := "SELECT DoctorID, Name, Email FROM Doctors WHERE DoctorID = ?"
	err := db.QueryRow(query, request.DoctorID).Scan(&doctor.DoctorID, &doctor.Name, &doctor.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Doctor not found", http.StatusNotFound)
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
	if err := json.NewEncoder(w).Encode(doctor); err != nil {
		log.Println("JSON encoding error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
