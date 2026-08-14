package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"oryoo.com/helper"
	"oryoo.com/models"
)

func FeatureRequestCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	CreateFeatureRequest(w, r)
}

func FeatureRequestCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ListFeatureRequests(w, r)
}

func FeatureRequestByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/feature-requests/")
	id = strings.Trim(id, "/")
	if id == "" || id == "create" {
		http.Error(w, "Feature request ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetFeatureRequest(w, r, id)
	case http.MethodPut, http.MethodPatch:
		UpdateFeatureRequest(w, r, id)
	case http.MethodDelete:
		DeleteFeatureRequest(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func featureRequestOwner(r *http.Request, bodyCreatedBy string) string {
	return marketingSpendOwner(r, bodyCreatedBy)
}

func requireFeatureRequestOwner(w http.ResponseWriter, r *http.Request) (string, bool) {
	owner := featureRequestOwner(r, "")
	if owner == "" {
		http.Error(w, "created_by is required", http.StatusBadRequest)
		return "", false
	}
	return owner, true
}

func CreateFeatureRequest(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFeatureRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.CreatedBy = featureRequestOwner(r, req.CreatedBy)

	record, err := models.ResolveCreateFeatureRequest(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record.ID = uuid.New().String()

	if err := helper.InsertFeatureRequest(record); err != nil {
		http.Error(w, "Failed to create feature request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Feature request created successfully",
		"feature_request": record,
	})
}

func ListFeatureRequests(w http.ResponseWriter, r *http.Request) {
	owner, ok := requireFeatureRequestOwner(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	items, err := helper.ListFeatureRequests(helper.FeatureRequestListFilter{
		CreatedBy: owner,
		Category:  q.Get("category"),
		Status:    q.Get("status"),
	})
	if err != nil {
		http.Error(w, "Failed to fetch feature requests", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":          true,
		"count":            len(items),
		"feature_requests": items,
	})
}

func GetFeatureRequest(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireFeatureRequestOwner(w, r)
	if !ok {
		return
	}

	record, err := helper.GetFeatureRequestByID(id)
	if err != nil {
		if err.Error() == "feature request not found" {
			http.Error(w, "Feature request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch feature request", http.StatusInternalServerError)
		return
	}
	if record.CreatedBy != owner {
		http.Error(w, "Feature request not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"feature_request": record,
	})
}

func UpdateFeatureRequest(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireFeatureRequestOwner(w, r)
	if !ok {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	var req models.UpdateFeatureRequestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	present := map[string]bool{
		"title":       models.JSONFieldPresent(raw, "title"),
		"description": models.JSONFieldPresent(raw, "description"),
		"category":    models.JSONFieldPresent(raw, "category"),
	}

	existing, err := helper.GetFeatureRequestByID(id)
	if err != nil {
		if err.Error() == "feature request not found" {
			http.Error(w, "Feature request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch feature request", http.StatusInternalServerError)
		return
	}
	if existing.CreatedBy != owner {
		http.Error(w, "Feature request not found", http.StatusNotFound)
		return
	}

	if err := models.ApplyFeatureRequestUpdate(existing, &req, present); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := helper.UpdateFeatureRequest(existing); err != nil {
		if err.Error() == "feature request not found" {
			http.Error(w, "Feature request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update feature request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Feature request updated successfully",
		"feature_request": existing,
	})
}

func DeleteFeatureRequest(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireFeatureRequestOwner(w, r)
	if !ok {
		return
	}

	if err := helper.DeleteFeatureRequest(id, owner); err != nil {
		if err.Error() == "feature request not found" {
			http.Error(w, "Feature request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete feature request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Feature request deleted successfully",
		"id":      id,
	})
}
