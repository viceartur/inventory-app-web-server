package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"inv_app/services/email"
)

func SendEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req email.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := email.SendEmail(req); err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"})
}
