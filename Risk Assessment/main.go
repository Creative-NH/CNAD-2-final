package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	/*
		// Get local port from environment variables
		localPort := os.Getenv("LOCAL_PORT")
		if localPort == "" {
			log.Fatal("LOCAL_PORT environment variable is not set")
		}
	*/

	// Initialize the router
	router := mux.NewRouter()
	router.HandleFunc("/api/analyzeRisk", func(w http.ResponseWriter, r *http.Request) {
		analyzeRiskHandler(w, r)
	}).Methods("POST")

	// Enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler.Handler(router)))
}

// Get recommendations based on risk level
func getRecommendation(riskLevel string) string {
	switch riskLevel {
	case "Low":
		return "Maintain a healthy lifestyle with balance exercises and check-ups."
	case "Moderate":
		return "Consider physical therapy, improve home safety, and monitor medications."
	case "High":
		return "Consult a healthcare provider for a fall risk assessment and use mobility aids."
	default:
		return "No recommendation available."
	}
}

// Analyze risk
func analyzeRiskHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		UserID  int         `json:"user_id"`
		Answers map[int]int `json:"answers"` // {question_id: selected_option_index (1-based)}
	}

	var req Request

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	// Validate User ID
	if req.UserID <= 0 {
		http.Error(w, "Invalid or missing user_id", http.StatusBadRequest)
		return
	}

	// Ensure at least some answers are provided
	if len(req.Answers) == 0 {
		http.Error(w, "No answers provided", http.StatusBadRequest)
		return
	}

	// Define risk points (Now indexed from 1)
	riskPoints := map[int]map[int]int{
		1:  {1: 2, 2: 0},             // Dizziness: Yes = 2, No = 0
		2:  {1: 1, 2: 2, 3: 3},       // Balance: Good = 1, Moderate = 2, Poor = 3
		3:  {1: 1, 2: 2, 3: 3},       // Falls: 0 = 1, 1-2 = 2, 3+ = 3
		4:  {1: 0, 2: 1, 3: 2, 4: 3}, // Mobility aid
		5:  {1: 0, 2: 1, 3: 2, 4: 3}, // Unsteady Walking
		6:  {1: 2, 2: 0},             // Recent fall: Yes = 2, No = 0
		7:  {1: 0, 2: 2},             // Stand without hands
		8:  {1: 2, 2: 0, 3: 1},       // Medications
		9:  {1: 0, 2: 2},             // Exercise
		10: {1: 2, 2: 0, 3: 1},       // Numbness
	}

	// Compute total score
	totalScore := 0
	for questionID, answerIndex := range req.Answers {
		if pointsMap, exists := riskPoints[questionID]; exists {
			if points, valid := pointsMap[answerIndex]; valid {
				totalScore += points
			} else {
				log.Printf("Invalid answer index %d for question %d", answerIndex, questionID)
			}
		} else {
			log.Printf("Invalid question ID: %d", questionID)
		}
	}

	// Determine risk level
	var riskLevel string
	switch {
	case totalScore <= 5:
		riskLevel = "Low"
	case totalScore <= 10:
		riskLevel = "Moderate"
	default:
		riskLevel = "High"
	}

	// Generate recommendation
	recommendation := getRecommendation(riskLevel)

	// Send JSON response
	response := map[string]interface{}{
		"total_score":    totalScore,
		"risk_level":     riskLevel,
		"recommendation": recommendation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
