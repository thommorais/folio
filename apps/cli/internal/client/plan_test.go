package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlanInputSendsTheServersIssueKey(t *testing.T) {
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"id":"p1","issue_id":"tk1","title":"Ship","status":"draft","tags":[]}`))
	}))
	defer server.Close()

	ticket := "tk1"
	plan, err := New(server.URL, "tok").UpdatePlan("p1", PlanInput{TicketID: &ticket})
	if err != nil {
		t.Fatalf("UpdatePlan() error = %v", err)
	}
	if got["issue_id"] != "tk1" {
		t.Errorf("body = %v, want issue_id tk1", got)
	}
	if plan.TicketID != "tk1" {
		t.Errorf("TicketID = %q, want tk1 read back from issue_id", plan.TicketID)
	}
}

func TestListPlansFiltersByTicket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"plans":[{"id":"p1","issue_id":"tk1","title":"A","tags":[]},{"id":"p2","issue_id":"tk2","title":"B","tags":[]},{"id":"p3","title":"C","tags":[]}]}`))
	}))
	defer server.Close()

	plans, err := New(server.URL, "tok").ListPlans("folio", PlanFilter{TicketID: "tk1"})
	if err != nil {
		t.Fatalf("ListPlans() error = %v", err)
	}
	if len(plans) != 1 || plans[0].ID != "p1" {
		t.Fatalf("plans = %+v, want only p1", plans)
	}
}
