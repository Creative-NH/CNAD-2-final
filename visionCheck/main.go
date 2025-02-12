package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type VisionResult struct {
	UserID        int    `json:"UserID"`
	LeftEyeScore  int    `json:"LeftEyeScore"`
	RightEyeScore int    `json:"RightEyeScore"`
	Comments      string `json:"Comments"`
	CreatedAt     string `json:"CreatedAt"` // Added field
}

func main() {
	http.HandleFunc("/postVisionResult", handlePostRequest)
	http.HandleFunc("/getLatestResult", getLatestResult)
	http.HandleFunc("/getAllVisionResults", getAllVisionResults)

	log.Println("Vision service running on port 8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var result VisionResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:04D685362v98@tcp(127.0.0.1:3306)/vision_assessment_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `INSERT INTO visionResults (UserID, LeftEyeScore, RightEyeScore, Comments) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, result.UserID, result.LeftEyeScore, result.RightEyeScore, result.Comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call Email Microservice if vision score is low
	if result.LeftEyeScore <= 2 || result.RightEyeScore <= 2 {
		go callEmailMicroservice(result)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Results stored successfully"))
}

func getLatestResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:04D685362v98@tcp(127.0.0.1:3306)/vision_assessment_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var result VisionResult
	query := `SELECT UserID, LeftEyeScore, RightEyeScore, Comments, CreatedAt FROM visionResults WHERE UserID = ? ORDER BY CreatedAt DESC LIMIT 1`
	err = db.QueryRow(query, userID).Scan(&result.UserID, &result.LeftEyeScore, &result.RightEyeScore, &result.Comments, &result.CreatedAt) // Added CreatedAt
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No results found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// All vision test results for the given userID
func getAllVisionResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:04D685362v98@tcp(127.0.0.1:3306)/vision_assessment_db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `SELECT UserID, LeftEyeScore, RightEyeScore, Comments, CreatedAt FROM visionResults WHERE UserID = ? ORDER BY CreatedAt DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []VisionResult
	for rows.Next() {
		var result VisionResult
		err := rows.Scan(&result.UserID, &result.LeftEyeScore, &result.RightEyeScore, &result.Comments, &result.CreatedAt) // Now scanning CreatedAt directly
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Call Email Microservice
func callEmailMicroservice(result VisionResult) {
	emailServiceURL := "http://localhost:8090/sendReportToDoctor"

	// Convert result to JSON
	requestBody, _ := json.Marshal(result)

	// Send POST request to email microservice
	resp, err := http.Post(emailServiceURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Error sending report to email microservice:", err)
		return
	}
	defer resp.Body.Close()

	log.Println("Report successfully sent to email microservice")
}
