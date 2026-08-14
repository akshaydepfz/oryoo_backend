package models

import (
	"fmt"
	"strings"
	"time"
)

const FeatureRequestStatusPending = "pending"

var FeatureRequestStatuses = map[string]bool{
	"pending":   true,
	"reviewing": true,
	"planned":   true,
	"completed": true,
	"rejected":  true,
}

type FeatureRequest struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    *string    `json:"category"`
	CreatedBy   string     `json:"created_by"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type CreateFeatureRequestRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    *string `json:"category"`
	CreatedBy   string  `json:"created_by"`
	Status      string  `json:"status"` // ignored; always stored as pending
}

type UpdateFeatureRequestRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Status      *string `json:"status"`     // ignored for user APIs
	CreatedBy   *string `json:"created_by"` // ignored for user APIs
}

func ValidateFeatureRequestFields(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description cannot be empty")
	}
	return nil
}

func ResolveCreateFeatureRequest(req *CreateFeatureRequestRequest) (*FeatureRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	if err := ValidateFeatureRequestFields(title, description); err != nil {
		return nil, err
	}
	createdBy := strings.TrimSpace(req.CreatedBy)
	if createdBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	return &FeatureRequest{
		Title:       title,
		Description: description,
		Category:    NormalizeOptionalString(req.Category),
		CreatedBy:   createdBy,
		Status:      FeatureRequestStatusPending,
	}, nil
}

func ApplyFeatureRequestUpdate(existing *FeatureRequest, req *UpdateFeatureRequestRequest, present map[string]bool) error {
	if existing == nil {
		return fmt.Errorf("feature request not found")
	}
	if req == nil {
		return nil
	}

	title := existing.Title
	description := existing.Description

	if present["title"] {
		if req.Title == nil {
			return fmt.Errorf("title cannot be empty")
		}
		title = strings.TrimSpace(*req.Title)
	}
	if present["description"] {
		if req.Description == nil {
			return fmt.Errorf("description cannot be empty")
		}
		description = strings.TrimSpace(*req.Description)
	}
	if err := ValidateFeatureRequestFields(title, description); err != nil {
		return err
	}

	existing.Title = title
	existing.Description = description
	if present["category"] {
		existing.Category = NormalizeOptionalString(req.Category)
	}
	return nil
}

func IsValidFeatureRequestStatus(status string) bool {
	return FeatureRequestStatuses[strings.ToLower(strings.TrimSpace(status))]
}
