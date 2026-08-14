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

func MarketingSpendCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	CreateMarketingSpend(w, r)
}

func MarketingSpendSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	GetMarketingSpendSummary(w, r)
}

func MarketingSpendCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ListMarketingSpend(w, r)
}

func MarketingSpendByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/marketing-spend/")
	id = strings.Trim(id, "/")
	if id == "" || id == "create" || id == "summary" {
		http.Error(w, "Marketing spend ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetMarketingSpend(w, r, id)
	case http.MethodPut, http.MethodPatch:
		UpdateMarketingSpend(w, r, id)
	case http.MethodDelete:
		DeleteMarketingSpend(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func marketingSpendOwner(r *http.Request, bodyCreatedBy string) string {
	if s := strings.TrimSpace(bodyCreatedBy); s != "" {
		return s
	}
	if s := strings.TrimSpace(r.URL.Query().Get("created_by")); s != "" {
		return s
	}
	if s := strings.TrimSpace(r.Header.Get("X-Firebase-UID")); s != "" {
		return s
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(auth, "Bearer ") {
		uid := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if uid != "" && uid != adminSecret {
			return uid
		}
	}
	return ""
}

func requireMarketingSpendOwner(w http.ResponseWriter, r *http.Request) (string, bool) {
	owner := marketingSpendOwner(r, "")
	if owner == "" {
		http.Error(w, "created_by is required", http.StatusBadRequest)
		return "", false
	}
	return owner, true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func CreateMarketingSpend(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMarketingSpendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.CreatedBy = marketingSpendOwner(r, req.CreatedBy)

	record, err := models.ResolveCreateMarketingSpend(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record.ID = uuid.New().String()

	if err := helper.InsertMarketingSpend(record); err != nil {
		http.Error(w, "Failed to create marketing spend", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Marketing spend created successfully",
		"marketing_spend": record,
	})
}

func ListMarketingSpend(w http.ResponseWriter, r *http.Request) {
	owner, ok := requireMarketingSpendOwner(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	items, err := helper.ListMarketingSpend(helper.MarketingSpendListFilter{
		CreatedBy: owner,
		Platform:  q.Get("platform"),
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
		Month:     q.Get("month"),
	})
	if err != nil {
		http.Error(w, "Failed to fetch marketing spend", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":          true,
		"count":            len(items),
		"marketing_spends": items,
	})
}

func GetMarketingSpend(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireMarketingSpendOwner(w, r)
	if !ok {
		return
	}

	record, err := helper.GetMarketingSpendByID(id)
	if err != nil {
		if err.Error() == "marketing spend not found" {
			http.Error(w, "Marketing spend not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch marketing spend", http.StatusInternalServerError)
		return
	}
	if record.CreatedBy != owner {
		http.Error(w, "Marketing spend not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"marketing_spend": record,
	})
}

func UpdateMarketingSpend(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireMarketingSpendOwner(w, r)
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
	var req models.UpdateMarketingSpendRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	present := map[string]bool{
		"platform":          models.JSONFieldPresent(raw, "platform"),
		"campaign_name":     models.JSONFieldPresent(raw, "campaign_name"),
		"amount_spent":      models.JSONFieldPresent(raw, "amount_spent"),
		"start_date":        models.JSONFieldPresent(raw, "start_date"),
		"end_date":          models.JSONFieldPresent(raw, "end_date"),
		"orders_generated":  models.JSONFieldPresent(raw, "orders_generated"),
		"revenue_generated": models.JSONFieldPresent(raw, "revenue_generated"),
		"notes":             models.JSONFieldPresent(raw, "notes"),
	}

	existing, err := helper.GetMarketingSpendByID(id)
	if err != nil {
		if err.Error() == "marketing spend not found" {
			http.Error(w, "Marketing spend not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch marketing spend", http.StatusInternalServerError)
		return
	}
	if existing.CreatedBy != owner {
		http.Error(w, "Marketing spend not found", http.StatusNotFound)
		return
	}

	if err := models.ApplyMarketingSpendUpdate(existing, &req, present); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := helper.UpdateMarketingSpend(existing); err != nil {
		if err.Error() == "marketing spend not found" {
			http.Error(w, "Marketing spend not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update marketing spend", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Marketing spend updated successfully",
		"marketing_spend": existing,
	})
}

func DeleteMarketingSpend(w http.ResponseWriter, r *http.Request, id string) {
	owner, ok := requireMarketingSpendOwner(w, r)
	if !ok {
		return
	}

	if err := helper.DeleteMarketingSpend(id, owner); err != nil {
		if err.Error() == "marketing spend not found" {
			http.Error(w, "Marketing spend not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete marketing spend", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Marketing spend deleted successfully",
		"id":      id,
	})
}

func GetMarketingSpendSummary(w http.ResponseWriter, r *http.Request) {
	owner, ok := requireMarketingSpendOwner(w, r)
	if !ok {
		return
	}

	summary, err := helper.GetMarketingSpendSummary(owner)
	if err != nil {
		http.Error(w, "Failed to fetch marketing spend summary", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
