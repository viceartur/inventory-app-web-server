package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"inv_app/services/email"

	"github.com/gorilla/mux"
)

func EmailInventoryReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId, _ := strconv.Atoi(vars["customerId"])

	var req email.EmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := email.EmailInventoryReport(req, customerId); err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"})
}
