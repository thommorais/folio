package pb_test

import (
	"slices"
	"testing"

	"github.com/pocketbase/pocketbase/core"

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
		{pb.ColIssues, "resolution"},
		{pb.ColIssues, "resolution_entry"},
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

func TestEntriesAcceptTheResolutionKind(t *testing.T) {
	s := setup(t)

	c, err := s.app.FindCollectionByNameOrId(pb.ColEntries)
	if err != nil {
		t.Fatal(err)
	}
	field, ok := c.Fields.GetByName("kind").(*core.SelectField)
	if !ok {
		t.Fatal("entries.kind is not a select field")
	}
	field.Values = slices.DeleteFunc(field.Values, func(v string) bool { return v == "resolution" })
	if err := s.app.Save(c); err != nil {
		t.Fatal(err)
	}

	if err := pb.Register(s.app); err != nil {
		t.Fatal(err)
	}

	c, err = s.app.FindCollectionByNameOrId(pb.ColEntries)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(c.Fields.GetByName("kind").(*core.SelectField).Values, "resolution") {
		t.Fatal("Register did not add the resolution kind to an existing collection")
	}
}
