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

	if err := email.EmailInventoryReport(req, customerId); err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}

	res := SuccessResponse{Message: "Email sent successfully."}
	json.NewEncoder(w).Encode(res)
}
