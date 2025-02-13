package main

import (
	"bytes"
	"crypto/tls"
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
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"gopkg.in/gomail.v2"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file:", err)
	}

	// Get database connection details from environment variables
	dbUser, dbPassword, dbHost, dbName := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME")
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" {
		log.Fatal("Database credentials not fully set in environment variables")
	}

	// Get local port from environment variables
	//localPort := os.Getenv("LOCAL_PORT")
	localPort := "5000"
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
	router.HandleFunc("/api/questionnaire", func(w http.ResponseWriter, r *http.Request) {
		questionnaireHandler(w, r, db)
	}).Methods("GET")
	router.HandleFunc("/api/addAssessmentResults", func(w http.ResponseWriter, r *http.Request) {
		addAssessmentHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/getLastAssessment", func(w http.ResponseWriter, r *http.Request) {
		getLastAssessmentHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/getAssessment", func(w http.ResponseWriter, r *http.Request) {
		getAssessmentHandler(w, r, db)
	}).Methods("POST")
	router.HandleFunc("/api/assessmentHistory", func(w http.ResponseWriter, r *http.Request) {
		assessmentHistoryHandler(w, r, db)
	}).Methods("POST")

	// CORS Configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Printf("Starting server on :%s", localPort)
	if err := http.ListenAndServe(":"+localPort, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Structure to handle assessment submissions

type Assessment struct {
	AssessmentID   int    `json:"id,omitempty"`
	TotalScore     int    `json:"totalScore"`
	RiskLevel      string `json:"riskLevel"`
	Recommendation string `json:"recommendation"`
	DateCreated    string `json:"dateCreated,omitempty"`
	UserID         int    `json:"user_id,omitempty"`
}

// Handler to retrieve questionnaire questions based on language (GET request with query string)
func questionnaireHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get language from query parameters
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "English" // Default to English if not specified
	}

	// Map supported languages to their respective tables
	languageMap := map[string]string{
		"English": "QuestionsEn",
		"Chinese": "QuestionsCn",
		"Malay":   "QuestionsMy",
		"Tamil":   "QuestionsTa",
	}

	// Check if language is supported
	tableName, exists := languageMap[language]
	if !exists {
		http.Error(w, "Unsupported language", http.StatusBadRequest)
		return
	}

	// Query the database for questions in the selected language
	query := fmt.Sprintf("SELECT QuestionID, QuestionContent, QuestionOptions FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Store retrieved questions
	var questions []map[string]interface{}

	for rows.Next() {
		var questionID int
		var questionContent string
		var questionOptions string // Stored in JSON format

		if err := rows.Scan(&questionID, &questionContent, &questionOptions); err != nil {
			log.Println("Data retrieval error: ", err)
			http.Error(w, "Data retrieval error", http.StatusInternalServerError)
			return
		}

		questions = append(questions, map[string]interface{}{
			"question_id":      questionID,
			"question_content": questionContent,
			"question_options": json.RawMessage(questionOptions), // Ensure JSON format
		})
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// Call Alert Service to send user notification
func sendNotification(userID int, riskLevel string) {
	// Define the notification message
	var message string
	if riskLevel == "Moderate" {
		message = "Your assessment indicates a moderate risk. Please follow the recommended exercises."
	} else if riskLevel == "High" {
		message = "High risk detected. Please contact your healthcare provider immediately."
	} else {
		return // No need to send notification for low risk
	}

	// Create JSON payload
	notificationBody, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"message": message,
	})

	// Send POST request to Alerts Service
	resp, err := http.Post("http://localhost:5002/api/postNotifications", "application/json", bytes.NewBuffer(notificationBody))
	if err != nil {
		log.Println("Failed to send notification:", err)
		return
	}
	defer resp.Body.Close()

	// Log notification response
	log.Printf("Notification sent for user %d with risk level %s\n", userID, riskLevel)
}

// Call Alert Service to send doctor alert
func sendAlertToDoctors(userID int, assessmentID int64) {
	log.Printf("Sending alert for high-risk user %d (Assessment ID: %d)\n", userID, assessmentID)

	// Create JSON payload with type "HealthAssessment"
	alertBody, _ := json.Marshal(map[string]interface{}{
		"assessment_id": assessmentID,
		"type":          "HealthAssessment", // Explicitly setting the type
	})

	// Send POST request to Notification Service
	resp, err := http.Post("http://localhost:5002/api/postAlerts", "application/json", bytes.NewBuffer(alertBody))
	if err != nil {
		log.Println("Failed to send alert to doctors:", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Alert successfully sent for assessment %d\n", assessmentID)
}

// Email sender function
func sendRiskAlertEmail(userID int, riskLevel string, assessmentID int64) {
	// Set up email details
	sender := "newuploadedvideo@gmail.com"
	password := "agof rvwb lreo tups"          // Use an App Password if using Gmail
	recipient := "s10247445@connect.np.edu.sg" // Change to the doctor's real email

	// Email subject and body
	subject := "Urgent: Risk Assessment Alert for User " + strconv.Itoa(userID)
	body := fmt.Sprintf(`
		<h2>Risk Assessment Alert</h2>
		<p><strong>User ID:</strong> %d</p>
		<p><strong>Risk Level:</strong> %s</p>
		<p>Please review the report immediately.</p>
		<p><a href="http://localhost:5500/report.html?assessment_id=%d" style="color: #007bff; font-weight: bold;">View Report</a></p>
	`, userID, riskLevel, assessmentID)

	// Configure SMTP
	d := gomail.NewDialer("smtp.gmail.com", 587, sender, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send risk alert email:", err)
	} else {
		log.Println("Risk alert email sent successfully to", recipient)
	}
}

// Add Results from Risk Assessment into DB
func addAssessmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	type Request struct {
		UserID  int         `json:"user_id"`
		Answers map[int]int `json:"answers"`
	}

	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding JSON request:", err)
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		log.Println("Invalid UserID:", req.UserID)
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Convert answers to JSON format
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		log.Println("Error marshalling answers JSON:", err)
		http.Error(w, "Failed to process answers", http.StatusInternalServerError)
		return
	}

	// Prepare JSON body for Risk Assessment Service
	riskRequestBody, err := json.Marshal(req)
	if err != nil {
		log.Println("Error encoding risk assessment request:", err)
		http.Error(w, "Failed to encode risk assessment request", http.StatusInternalServerError)
		return
	}

	// Call Risk Assessment Service
	riskResponse, err := http.Post("http://localhost:8080/api/analyzeRisk", "application/json", bytes.NewBuffer(riskRequestBody))
	if err != nil {
		log.Println("Error calling Risk Assessment Service:", err)
		http.Error(w, "Failed to process risk assessment", http.StatusInternalServerError)
		return
	}
	defer riskResponse.Body.Close()

	// Parse Risk Assessment Response
	var riskResult struct {
		TotalScore     int    `json:"total_score"`
		RiskLevel      string `json:"risk_level"`
		Recommendation string `json:"recommendation"`
	}

	if err := json.NewDecoder(riskResponse.Body).Decode(&riskResult); err != nil {
		log.Println("Error decoding risk assessment response:", err)
		http.Error(w, "Failed to parse risk assessment response", http.StatusInternalServerError)
		return
	}

	// Store results in the database
	insertQuery := `INSERT INTO Assessments (UserID, QuestionResponses, TotalScore, RiskLevel, Recommendation, DateCreated) 
                    VALUES (?, ?, ?, ?, ?, NOW())`
	result, err := db.Exec(insertQuery, req.UserID, answersJSON, riskResult.TotalScore, riskResult.RiskLevel, riskResult.Recommendation)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Failed to store assessment data", http.StatusInternalServerError)
		return
	}

	assessmentID, _ := result.LastInsertId()

	// If risk is MODERATE or HIGH, send email to doctor
	if riskResult.RiskLevel == "Moderate" || riskResult.RiskLevel == "High" {
		go sendRiskAlertEmail(req.UserID, riskResult.RiskLevel, assessmentID)
	}

	// If risk is MODERATE or HIGH, send a notification
	if riskResult.RiskLevel == "Moderate" || riskResult.RiskLevel == "High" {
		go sendNotification(req.UserID, riskResult.RiskLevel)
	}

	// **If risk is HIGH, send an alert to doctors**
	if riskResult.RiskLevel == "High" {
		go sendAlertToDoctors(req.UserID, assessmentID)
	}

	// Send response
	response := map[string]interface{}{
		"assessment_id":      assessmentID,
		"user_id":            req.UserID,
		"total_score":        riskResult.TotalScore,
		"risk_level":         riskResult.RiskLevel,
		"recommendation":     riskResult.Recommendation,
		"question_responses": req.Answers,
	}

	log.Println("Successfully stored assessment:", response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Retrieve results of last assessment
func getLastAssessmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	// Query the database for the latest risk assessment for the user
	query := `SELECT AssessmentID, TotalScore, RiskLevel, Recommendation 
              FROM Assessments 
              WHERE UserID = ? 
              ORDER BY DateCreated DESC LIMIT 1`

	assessment := Assessment{}

	err := db.QueryRow(query, req.UserID).Scan(
		&assessment.AssessmentID,
		&assessment.TotalScore,
		&assessment.RiskLevel,
		&assessment.Recommendation,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No risk assessments found for this user", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		}
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessment)
}

// Retrieve results of a specific assessment
func getAssessmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	// Query the database for the risk assessment for the user
	query := `SELECT TotalScore, RiskLevel, Recommendation, UserID
              FROM Assessments 
              WHERE AssessmentID = ?`

	assessment := Assessment{}

	err := db.QueryRow(query, req.AssessmentID).Scan(
		&assessment.TotalScore,
		&assessment.RiskLevel,
		&assessment.Recommendation,
		&assessment.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No risk assessments found for this user", http.StatusNotFound)
		} else {
			log.Println("Database query error:", err)
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		}
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessment)
}

func assessmentHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	// Query the database for all risk assessments for the user
	query := `SELECT AssessmentID, TotalScore, RiskLevel, Recommendation, DateCreated
              FROM Assessments 
              WHERE UserID = ? 
              ORDER BY DateCreated DESC`

	rows, err := db.Query(query, req.UserID)
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var assessments []Assessment

	// Iterate over rows
	for rows.Next() {
		var assessment Assessment
		var dateCreated time.Time

		if err := rows.Scan(
			&assessment.AssessmentID,
			&assessment.TotalScore,
			&assessment.RiskLevel,
			&assessment.Recommendation,
			&dateCreated,
		); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to process data", http.StatusInternalServerError)
			return
		}

		assessment.DateCreated = dateCreated.Format("2006-01-02 15:04:05")
		assessments = append(assessments, assessment)
	}

	// Check if no records were found
	if len(assessments) == 0 {
		http.Error(w, "No risk assessments found for this user", http.StatusNotFound)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assessments)
}
