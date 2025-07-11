package handlers

import (
	"context"
	"encoding/json"
	"inv_app/database"
	"inv_app/services/materials"
	"net/http"
	"strconv"
)

func GetMaterialTypesHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	materialTypes, err := materials.FetchMaterialTypes(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(materialTypes)
}

func GetMaterialUsageReasonsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	materialTypes, err := materials.FetchMaterialUsageReasons(db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(materialTypes)
}

func SendMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var material materials.IncomingMaterial
	json.NewDecoder(r.Body).Decode(&material)
	err := materials.SendMaterial(material, db)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(material)
}

func GetIncomingMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	materialId := r.URL.Query().Get("materialId")
	id, _ := strconv.Atoi(materialId)
	materials, err := materials.GetIncomingMaterials(db, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func UpdateIncomingMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var material materials.IncomingMaterial
	json.NewDecoder(r.Body).Decode(&material)
	err := materials.UpdateIncomingMaterial(db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Requested Material Updated"}
	json.NewEncoder(w).Encode(res)
}

func DeleteIncomingMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var material materials.IncomingMaterial
	json.NewDecoder(r.Body).Decode(&material)
	shippingId := material.ShippingID

	ctx := context.TODO()
	err := materials.DeleteIncomingMaterial(ctx, db, shippingId)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Incoming Material Deleted"}
	json.NewEncoder(w).Encode(res)
}

func CreateMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var material materials.MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	materialId, err := materials.CreateMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Material ID created", Data: materialId}
	json.NewEncoder(w).Encode(res)
}

func GetMaterialsLikeHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	materialId := r.URL.Query().Get("materialId")
	id, _ := strconv.Atoi(materialId)
	stockId := r.URL.Query().Get("stockId")
	programName := r.URL.Query().Get("programName")
	description := r.URL.Query().Get("description")
	locationName := r.URL.Query().Get("locationName")

	filterOpts := &materials.MaterialFilter{
		MaterialId:   id,
		StockId:      stockId,
		ProgramName:  programName,
		Description:  description,
		LocationName: locationName,
	}
	materials, err := materials.GetMaterials(db, filterOpts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func GetMaterialsExactHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	stockId := r.URL.Query().Get("stockId")
	filterOpts := &materials.MaterialFilter{
		StockId: stockId,
	}
	materials, err := materials.GetMaterialsByStockID(db, filterOpts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func UpdateMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var material materials.MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	err := materials.UpdateMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusNotFound)
		return
	}
	res := SuccessResponse{Message: "Material Updated"}
	json.NewEncoder(w).Encode(res)
}

func MoveMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var material materials.MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	err := materials.MoveMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Material Moved", Data: material}
	json.NewEncoder(w).Encode(res)
}

func RemoveMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var material materials.MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)

	ctx := context.TODO()
	err := materials.RemoveMaterial(ctx, db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Material Quantity Removed", Data: material}
	json.NewEncoder(w).Encode(res)
}

func RequestMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var materialsData materials.RequestedMaterialsJSON
	json.NewDecoder(r.Body).Decode(&materialsData)

	ctx := context.TODO()
	err := materials.RequestMaterials(ctx, db, materialsData)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Materials requested"}
	json.NewEncoder(w).Encode(res)
}

func GetRequestedMaterialsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	requestId := r.URL.Query().Get("requestId")
	id, _ := strconv.Atoi(requestId)
	stockId := r.URL.Query().Get("stockId")
	status := r.URL.Query().Get("status")
	requestedFrom := r.URL.Query().Get("requestedFrom")
	requestedTo := r.URL.Query().Get("requestedTo")
	filterOpts := materials.MaterialFilter{
		RequestId:     id,
		StockId:       stockId,
		Status:        status,
		RequestedFrom: requestedFrom,
		RequestedTo:   requestedTo,
	}
	materials, err := materials.GetRequestedMaterials(db, filterOpts)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Requested Materials List", Data: materials}
	json.NewEncoder(w).Encode(res)
}

func GetRequestedMaterialsCountHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	materialsCount, err := materials.GetRequestedMaterialsCount(db)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Requested Materials Count", Data: materialsCount}
	json.NewEncoder(w).Encode(res)
}

func UpdateRequestedMaterialHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	var material materials.MaterialJSON
	json.NewDecoder(r.Body).Decode(&material)
	err := materials.UpdateRequestedMaterial(db, material)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Requested Material Updated"}
	json.NewEncoder(w).Encode(res)
}

func GetMaterialDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	stockId := r.URL.Query().Get("stockId")
	description, err := materials.GetMaterialDescription(db, stockId)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Material Description Requested", Data: description}
	json.NewEncoder(w).Encode(res)
}

func GetMaterialTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()

	jobTicket := r.URL.Query().Get("jobTicket")
	transactions, err := materials.GetMaterialTransactions(db, jobTicket)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusConflict)
		return
	}
	res := SuccessResponse{Message: "Material Transactions Requested", Data: transactions}
	json.NewEncoder(w).Encode(res)
}
