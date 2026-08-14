package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oryoo.com/models"
)

func TestCreateFeatureRequest_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "empty title",
			body: `{"title":"","description":"desc","created_by":"uid-1"}`,
			want: "title cannot be empty",
		},
		{
			name: "empty description",
			body: `{"title":"Add WhatsApp","description":"  ","created_by":"uid-1"}`,
			want: "description cannot be empty",
		},
		{
			name: "missing created_by",
			body: `{"title":"Add WhatsApp","description":"Send updates"}`,
			want: "created_by is required",
		},
		{
			name: "invalid json",
			body: `{`,
			want: "Invalid JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/feature-requests/create", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			CreateFeatureRequest(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tc.want) {
				t.Fatalf("body=%s want %q", rr.Body.String(), tc.want)
			}
		})
	}
}

func TestCreateFeatureRequest_IgnoresClientStatus(t *testing.T) {
	var req models.CreateFeatureRequestRequest
	payload := `{"title":"Add WhatsApp","description":"Send updates","created_by":"uid-1","status":"completed"}`
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	record, err := models.ResolveCreateFeatureRequest(&req)
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != "pending" {
		t.Fatalf("status=%q", record.Status)
	}
}

func TestListFeatureRequests_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feature-requests", nil)
	rr := httptest.NewRecorder()
	ListFeatureRequests(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetFeatureRequest_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feature-requests/abc", nil)
	rr := httptest.NewRecorder()
	GetFeatureRequest(rr, req, "abc")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteFeatureRequest_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/feature-requests/abc", nil)
	rr := httptest.NewRecorder()
	DeleteFeatureRequest(rr, req, "abc")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestUpdateFeatureRequest_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/feature-requests/abc", strings.NewReader(`{"title":"x"}`))
	rr := httptest.NewRecorder()
	UpdateFeatureRequest(rr, req, "abc")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFeatureRequestOwnerPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feature-requests", nil)
	req.Header.Set("X-Firebase-UID", "uid-header")
	if got := featureRequestOwner(req, ""); got != "uid-header" {
		t.Fatalf("got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/feature-requests?created_by=uid-query", nil)
	if got := featureRequestOwner(req, ""); got != "uid-query" {
		t.Fatalf("got %q", got)
	}

	req = httptest.NewRequest(http.MethodPost, "/feature-requests/create", nil)
	if got := featureRequestOwner(req, "uid-body"); got != "uid-body" {
		t.Fatalf("got %q", got)
	}
}

func TestFeatureRequestByIDHandler_RejectsCreatePath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feature-requests/create", nil)
	rr := httptest.NewRecorder()
	FeatureRequestByIDHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFeatureRequestEmptyListJSON(t *testing.T) {
	items := make([]models.FeatureRequest, 0)
	body, err := json.Marshal(map[string]interface{}{
		"success":          true,
		"count":            len(items),
		"feature_requests": items,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"feature_requests":[]`) {
		t.Fatalf("%s", body)
	}
}
