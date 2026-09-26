package pb_test

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
)

func trySave(app core.App, collection string, values map[string]any) error {
	c, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		return err
	}
	r := core.NewRecord(c)
	for k, v := range values {
		r.Set(k, v)
	}
	return app.Save(r)
}

func TestClientMemberIsUniquePerUser(t *testing.T) {
	s := setup(t)

	row := map[string]any{"client": s.client.Id, "user": s.editor.Id, "role": "member"}
	newRecord(t, s.app, pb.ColClientMembers, row)

	row["role"] = "owner"
	if err := trySave(s.app, pb.ColClientMembers, row); err == nil {
		t.Error("a second membership row for the same user and client was accepted")
	}
}

func TestProjectGrantIsUniquePerTarget(t *testing.T) {
	s := setup(t)

	byDomain := map[string]any{"project": s.project.Id, "domain": s.other.Id, "role": "viewer"}
	byUser := map[string]any{"project": s.project.Id, "user": s.stranger.Id, "role": "viewer"}
	newRecord(t, s.app, pb.ColProjectGrants, byDomain)
	newRecord(t, s.app, pb.ColProjectGrants, byUser)

	byDomain["role"] = "editor"
	if err := trySave(s.app, pb.ColProjectGrants, byDomain); err == nil {
		t.Error("a second grant to the same domain was accepted")
	}
	byUser["role"] = "editor"
	if err := trySave(s.app, pb.ColProjectGrants, byUser); err == nil {
		t.Error("a second grant to the same user was accepted")
	}
}

func TestProjectGrantsToDifferentUsersCoexist(t *testing.T) {
	s := setup(t)

	first := newUser(t, s.app, "first@test.local")
	second := newUser(t, s.app, "second@test.local")
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "user": first.Id, "role": "editor"})
	if err := trySave(s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "user": second.Id, "role": "viewer"}); err != nil {
		t.Errorf("a grant to a second user was refused: %v", err)
	}
}

func TestBackfillSeedsClientOwnersAndDomainGrants(t *testing.T) {
	s := setup(t)

	for _, name := range []string{pb.ColClientMembers, pb.ColProjectGrants} {
		c, err := s.app.FindCollectionByNameOrId(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.app.Delete(c); err != nil {
			t.Fatalf("drop %s: %v", name, err)
		}
	}
	if err := pb.Register(s.app); err != nil {
		t.Fatal(err)
	}

	owners, err := s.app.FindAllRecords(pb.ColClientMembers)
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != 1 || owners[0].GetString("user") != s.owner.Id ||
		owners[0].GetString("client") != s.client.Id || owners[0].GetString("role") != "owner" {
		t.Errorf("client members = %v, want only the domain owner as client owner", owners)
	}

	grants, err := s.app.FindAllRecords(pb.ColProjectGrants)
	if err != nil {
		t.Fatal(err)
	}
	if len(grants) != 1 || grants[0].GetString("project") != s.project.Id ||
		grants[0].GetString("domain") != s.domain.Id || grants[0].GetString("user") != "" ||
		grants[0].GetString("role") != "editor" {
		t.Errorf("project grants = %v, want one editor grant from the project to its domain", grants)
	}
}
