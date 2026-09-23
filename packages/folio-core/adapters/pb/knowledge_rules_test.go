package pb_test

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
)

func knowledgeRecord(t *testing.T, s scenario) *core.Record {
	t.Helper()

	return newRecord(t, s.app, pb.ColKnowledge, map[string]any{
		"slug": "a-note", "title": "A note", "body": "worth keeping",
		"created_by": s.owner.Id,
	})
}

// Knowledge is the one collection a stranger may read. Every other collection
// hides a record from a non-member, so this is asserted rather than assumed:
// an accidental copy of the membership rule would be invisible otherwise.
func TestKnowledgeIsReadableByAnySignedInUser(t *testing.T) {
	s := setup(t)
	note := knowledgeRecord(t, s)

	for _, user := range []struct {
		name   string
		record *core.Record
	}{
		{"owner", s.owner},
		{"viewer", s.viewer},
		{"stranger", s.stranger},
	} {
		t.Run(user.name, func(t *testing.T) {
			if !canView(t, s.app, note, user.record) {
				t.Errorf("%s cannot read knowledge", user.name)
			}
		})
	}
}

// A viewer cannot write project content but can write knowledge: the roles
// that fence a project do not apply here.
func TestKnowledgeIsWritableByAnySignedInUser(t *testing.T) {
	s := setup(t)
	note := knowledgeRecord(t, s)
	rules := note.Collection()

	for _, user := range []struct {
		name   string
		record *core.Record
	}{
		{"viewer", s.viewer},
		{"stranger", s.stranger},
	} {
		t.Run(user.name, func(t *testing.T) {
			for _, rule := range []struct {
				name string
				expr *string
			}{
				{"create", rules.CreateRule},
				{"update", rules.UpdateRule},
				{"delete", rules.DeleteRule},
			} {
				ok, err := s.app.CanAccessRecord(note, info(user.record, "POST"), rule.expr)
				if err != nil {
					t.Fatalf("%s rule: %v", rule.name, err)
				}
				if !ok {
					t.Errorf("%s cannot %s knowledge", user.name, rule.name)
				}
			}
		})
	}
}

// Open to everyone still means everyone signed in: an unauthenticated request
// must not reach the collection at all.
func TestKnowledgeIsClosedToAnonymousRequests(t *testing.T) {
	s := setup(t)
	note := knowledgeRecord(t, s)
	rules := note.Collection()

	for _, rule := range []struct {
		name string
		expr *string
	}{
		{"view", rules.ViewRule},
		{"list", rules.ListRule},
		{"create", rules.CreateRule},
		{"update", rules.UpdateRule},
		{"delete", rules.DeleteRule},
	} {
		t.Run(rule.name, func(t *testing.T) {
			ok, err := s.app.CanAccessRecord(note, &core.RequestInfo{Method: "GET", Context: "default"}, rule.expr)
			if err != nil {
				t.Fatalf("%s rule: %v", rule.name, err)
			}
			if ok {
				t.Errorf("an anonymous request may %s knowledge", rule.name)
			}
		})
	}
}
