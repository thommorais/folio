package pb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	"folio/folio-core/domain"
)

// mapErr translates PocketBase storage errors into the domain sentinels the
// services and the HTTP adapter understand. PocketBase surfaces a missing
// record as sql.ErrNoRows from the underlying dbx query.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

// strSlice reads a JSON string-array field. A field that is empty, null or
// unparseable yields nil rather than an error: a malformed tag list should
// not make a record unreadable.
func strSlice(rec *core.Record, field string) []string {
	raw := rec.GetString(field)
	if raw == "" || raw == "null" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func jsonMap(rec *core.Record, field string) map[string]any {
	raw := rec.GetString(field)
	if raw == "" || raw == "null" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

// setJSON writes a slice or map, storing an empty value as an empty JSON
// array/object so reads stay symmetric.
func setJSON(rec *core.Record, field string, v any) {
	if v == nil {
		rec.Set(field, "[]")
		return
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		rec.Set(field, "[]")
		return
	}
	rec.Set(field, string(encoded))
}

// timePtr converts a PocketBase date, whose zero value means "unset", into a
// nil-able time.
func timePtr(dt types.DateTime) *time.Time {
	if dt.IsZero() {
		return nil
	}
	t := dt.Time()
	return &t
}

// setDate writes a nil-able time, clearing the field when nil.
func setDate(rec *core.Record, field string, t *time.Time) {
	if t == nil {
		rec.Set(field, "")
		return
	}
	dt, err := types.ParseDateTime(*t)
	if err != nil {
		rec.Set(field, "")
		return
	}
	rec.Set(field, dt)
}

// toIDs converts a stored JSON string array into typed todo IDs.
func toTodoIDs(raw []string) []domain.TodoID {
	if len(raw) == 0 {
		return nil
	}
	out := make([]domain.TodoID, 0, len(raw))
	for _, s := range raw {
		out = append(out, domain.TodoID(s))
	}
	return out
}

func fromTodoIDs(ids []domain.TodoID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func toTicketIDs(raw []string) []domain.TicketID {
	if len(raw) == 0 {
		return nil
	}
	out := make([]domain.TicketID, 0, len(raw))
	for _, s := range raw {
		out = append(out, domain.TicketID(s))
	}
	return out
}

func fromTicketIDs(ids []domain.TicketID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// isNotFound reports whether an already-mapped error means "absent".
func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}
