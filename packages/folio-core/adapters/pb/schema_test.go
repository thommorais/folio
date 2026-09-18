package pb_test

import (
	"testing"

	"folio/folio-core/adapters/pb"
)

// The repositories address these columns by name, so a relation dropped from
// the schema reads as an empty field rather than an error.
func TestSchemaKeepsTheRelationsTheRepositoriesRead(t *testing.T) {
	s := setup(t)

	for _, want := range []struct{ collection, field string }{
		{pb.ColPlans, "issue"},
		{pb.ColCycles, "issue"},
		{pb.ColEntries, "issue"},
		{pb.ColEntries, "plan"},
		{pb.ColEntries, "cycle"},
		{pb.ColIssues, "plan"},
		{pb.ColProjects, "domain"},
		{pb.ColMembers, "domain"},
		{pb.ColDomains, "client"},
	} {
		c, err := s.app.FindCollectionByNameOrId(want.collection)
		if err != nil {
			t.Fatalf("%s: %v", want.collection, err)
		}
		if c.Fields.GetByName(want.field) == nil {
			t.Errorf("%s has no %s field", want.collection, want.field)
		}
	}
}
