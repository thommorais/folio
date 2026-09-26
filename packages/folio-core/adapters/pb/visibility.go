package pb

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type idSet struct {
	seen map[string]bool
	ids  []any
}

func (s *idSet) add(id string) {
	if id == "" || s.seen[id] {
		return
	}
	if s.seen == nil {
		s.seen = map[string]bool{}
	}
	s.seen[id] = true
	s.ids = append(s.ids, id)
}

func column(rows []*core.Record, field string) []any {
	out := make([]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.GetString(field))
	}
	return out
}

func grantedDomains(app core.App, user string) ([]*core.Record, error) {
	grants, err := app.FindAllRecords(ColProjectGrants, dbx.HashExp{"user": user})
	if err != nil || len(grants) == 0 {
		return nil, err
	}
	projects, err := app.FindAllRecords(ColProjects, dbx.In("id", column(grants, "project")...))
	if err != nil || len(projects) == 0 {
		return nil, err
	}
	return app.FindAllRecords(ColDomains, dbx.In("id", column(projects, "domain")...))
}

func rosterDomains(app core.App, user string) ([]*core.Record, error) {
	rows, err := app.FindAllRecords(ColMembers, dbx.HashExp{"user": user})
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return app.FindAllRecords(ColDomains, dbx.In("id", column(rows, "domain")...))
}
