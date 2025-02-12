package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	localPort := "5002"
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
	router.HandleFunc("/api/getNotifications", func(w http.ResponseWriter, r *http.Request) {
		notificationHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/postNotifications", func(w http.ResponseWriter, r *http.Request) {
		postHandler(w, r, db)
	}).Methods("POST")

	// Alerts are for Doctors
	router.HandleFunc("/api/getAlerts", func(w http.ResponseWriter, r *http.Request) {
		doctorNotificationHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/postAlerts", func(w http.ResponseWriter, r *http.Request) {
		doctorPostHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/resolveAlerts/{assessment_id}", func(w http.ResponseWriter, r *http.Request) {
		doctorResolveHandler(w, r, db)
	}).Methods("DELETE")

	// CORS Configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Printf("Starting server on :%s", localPort)
	if err := http.ListenAndServe(":"+localPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func notificationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID int `json:"user_id"`
	}
	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid JSON request")
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		log.Println("Invalid or missing User ID")
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Query database for user notifications
	query := `SELECT NotificationID, Message, SentAt FROM Notifications WHERE UserID = ? ORDER BY SentAt DESC`
	rows, err := db.Query(query, req.UserID)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}

	// Iterate over rows
	for rows.Next() {
		var notificationID int
		var message string
		var sentAt time.Time

		if err := rows.Scan(&notificationID, &message, &sentAt); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to process data", http.StatusInternalServerError)
			return
		}

		notifications = append(notifications, map[string]interface{}{
			"notification_id": notificationID,
			"message":         message,
			"sent_at":         sentAt.Format("2006-01-02 15:04:05"),
		})
	}

	// If no notifications are found, return an empty JSON array []
	if len(notifications) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func postHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID  int    `json:"user_id"`
		Message string `json:"message"`
	}
	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid JSON request")
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.UserID <= 0 || req.Message == "" {
		log.Println("Invalid input: UserID or Message missing")
		http.Error(w, "Invalid input: UserID or Message missing", http.StatusBadRequest)
		return
	}

	// Insert into database
	query := `INSERT INTO Notifications (UserID, Message) VALUES (?, ?)`
	_, err := db.Exec(query, req.UserID, req.Message)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Failed to store notification", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]string{"message": "Notification sent successfully!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func doctorNotificationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Query database for alerts
	query := `SELECT AlertID, AssessmentID, SentAt FROM Alerts WHERE SentAt >= NOW() - INTERVAL 3 DAY ORDER BY SentAt DESC`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Failed to fetch alerts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []map[string]interface{}

	// Iterate over rows
	for rows.Next() {
		var alertID, assessmentID int
		var sentAt time.Time

		if err := rows.Scan(&alertID, &assessmentID, &sentAt); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to process data", http.StatusInternalServerError)
			return
		}

		alerts = append(alerts, map[string]interface{}{
			"alert_id":      alertID,
			"assessment_id": assessmentID,
			"sent_at":       sentAt.Format("2006-01-02 15:04:05"),
		})
	}

	// If no alerts found, return empty JSON array []
	if len(alerts) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func doctorPostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		AssessmentID int `json:"assessment_id"`
	}
	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid JSON request")
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.AssessmentID <= 0 {
		log.Println("Invalid input: AssessmentID missing")
		http.Error(w, "Invalid input: AssessmentID missing", http.StatusBadRequest)
		return
	}

	// Insert into database
	query := `INSERT INTO Alerts (AssessmentID) VALUES (?)`
	_, err := db.Exec(query, req.AssessmentID)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Failed to store alert", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]string{"message": "Alert sent successfully!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Resolve Alert - Removes Alert from DB
func doctorResolveHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse assessment_id from URL query parameters
	vars := mux.Vars(r)
	assessmentIDStr := vars["assessment_id"]

	// Validate input
	if assessmentIDStr == "" {
		log.Println("Invalid input: AssessmentID missing")
		http.Error(w, "Invalid input: AssessmentID missing", http.StatusBadRequest)
		return
	}

	// Convert assessmentID to integer
	assessmentID, err := strconv.Atoi(assessmentIDStr)
	if err != nil || assessmentID <= 0 {
		log.Println("Invalid input: AssessmentID must be a positive integer")
		http.Error(w, "Invalid input: AssessmentID must be a positive integer", http.StatusBadRequest)
		return
	}

	// Delete the alert from the database
	query := `DELETE FROM Alerts WHERE AssessmentID = ?`
	result, err := db.Exec(query, assessmentID)
	if err != nil {
		log.Println("Database delete error:", err)
		http.Error(w, "Failed to resolve alert", http.StatusInternalServerError)
		return
	}

	// Check if any row was deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No alert found for the given assessment ID", http.StatusNotFound)
		return
	}

	// Return success response
	response := map[string]string{"message": "Alert resolved successfully!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
