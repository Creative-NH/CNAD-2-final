package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gopkg.in/gomail.v2"
)

type VisionResult struct {
	UserID        int    `json:"UserID"`
	LeftEyeScore  int    `json:"LeftEyeScore"`
	RightEyeScore int    `json:"RightEyeScore"`
	Comments      string `json:"Comments"`
}

// Function to send email using Gmail SMTP
// Function to send email using Gmail SMTP
func sendEmailToDoctor(result VisionResult) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "newuploadedvideo@gmail.com") // Your sender email
	m.SetHeader("To", "s10247445@connect.np.edu.sg")  // Doctor's email
	m.SetHeader("Subject", "Urgent: Vision Test Report for User ID "+strconv.Itoa(result.UserID))

	// Email body with hyperlink to doctorHome.html
	body := fmt.Sprintf(`
		<h2>Vision Test Report</h2>
		<p><strong>User ID:</strong> %d</p>
		<p><strong>Left Eye Score:</strong> %d</p>
		<p><strong>Right Eye Score:</strong> %d</p>
		<p><strong>Comments:</strong> %s</p>
		<p>Please review the report and advise accordingly.</p>
		<p><a href="http://localhost:5500/doctorLogin.html" style="color: #007bff; font-weight: bold;">Click here to view the report</a></p>
	`, result.UserID, result.LeftEyeScore, result.RightEyeScore, result.Comments)

	m.SetBody("text/html", body)

	// Configure Gmail SMTP settings
	d := gomail.NewDialer("smtp.gmail.com", 587, "newuploadedvideo@gmail.com", "agof rvwb lreo tups")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // Needed for Gmail

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// API Endpoint to receive and send emails
func handleSendReportToDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var result VisionResult
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Send email
	err = sendEmailToDoctor(result)
	if err != nil {
		log.Println("Failed to send email:", err)
		http.Error(w, "Failed to send report to doctor", http.StatusInternalServerError)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Report sent successfully to doctor"})
}

func main() {
	http.HandleFunc("/sendReportToDoctor", handleSendReportToDoctor)
	log.Println("Email microservice running on port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil)) // Running on port 8090
}
