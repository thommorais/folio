package pb_test

import (
	"testing"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func TestCycleKeepsItsMap(t *testing.T) {
	s := setup(t)
	repo := pb.NewCycleRepository(s.app)
	theMap := newIssue(t, s, domain.IssueTicket, "plan-the-view")

	created, err := repo.Create(t.Context(), domain.Cycle{
		ID: "cyclemap0000001", ProjectID: domain.ProjectID(s.project.Id), IssueID: domain.IssueID(s.ticket.Id),
		Ordinal: 1, Phase: domain.PhasePlan, MapID: theMap.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.MapID != theMap.ID {
		t.Fatalf("map = %q, want %q", got.MapID, theMap.ID)
	}
}
