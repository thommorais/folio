package pb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type InterviewRepository struct {
	app core.App
}

func NewInterviewRepository(app core.App) *InterviewRepository {
	return &InterviewRepository{app: app}
}

var _ ports.InterviewRepository = (*InterviewRepository)(nil)

func toInterview(rec *core.Record) (domain.Interview, error) {
	var state domain.InterviewState
	if raw := rec.GetString("state"); raw != "" && raw != "null" {
		if err := json.Unmarshal([]byte(raw), &state); err != nil {
			return domain.Interview{}, fmt.Errorf("interview %s: read state: %w", rec.Id, err)
		}
	}
	return domain.Interview{
		ID:          domain.InterviewID(rec.Id),
		ProjectID:   domain.ProjectID(rec.GetString("project")),
		IssueID:     domain.IssueID(rec.GetString("issue")),
		Topic:       rec.GetString("topic"),
		State:       state,
		AgentStatus: domain.AgentStatus(rec.GetString("agent_status")),
		AgentSince:  rec.GetDateTime("agent_since").Time(),
		Handled:     rec.GetInt("handled"),
		FinishedAt:  timePtr(rec.GetDateTime("finished_at")),
		CreatedBy:   domain.UserID(rec.GetString("created_by")),
		CreatedAt:   rec.GetDateTime("created").Time(),
		UpdatedAt:   rec.GetDateTime("updated").Time(),
	}, nil
}

func applyInterview(rec *core.Record, i domain.Interview) {
	rec.Set("project", string(i.ProjectID))
	rec.Set("issue", string(i.IssueID))
	rec.Set("topic", i.Topic)
	rec.Set("state", i.State)
	rec.Set("agent_status", string(i.AgentStatus))
	setDate(rec, "agent_since", &i.AgentSince)
	rec.Set("handled", i.Handled)
	setDate(rec, "finished_at", i.FinishedAt)
	if i.CreatedBy != "" {
		rec.Set("created_by", string(i.CreatedBy))
	}
}

func (r *InterviewRepository) Create(ctx context.Context, i domain.Interview) (domain.Interview, error) {
	collection, err := r.app.FindCollectionByNameOrId(ColInterviews)
	if err != nil {
		return domain.Interview{}, mapErr(err)
	}
	rec := core.NewRecord(collection)
	if i.ID != "" {
		rec.Id = string(i.ID)
	}
	applyInterview(rec, i)
	if err := r.app.Save(rec); err != nil {
		return domain.Interview{}, mapErr(err)
	}
	return toInterview(rec)
}

func (r *InterviewRepository) GetByID(ctx context.Context, id domain.InterviewID) (domain.Interview, error) {
	rec, err := r.app.FindRecordById(ColInterviews, string(id))
	if err != nil {
		return domain.Interview{}, mapErr(err)
	}
	return toInterview(rec)
}

func (r *InterviewRepository) ListByIssue(ctx context.Context, issue domain.IssueID) ([]domain.Interview, error) {
	records, err := r.app.FindRecordsByFilter(ColInterviews, "issue = {:issue}", "created", 0, 0, dbx.Params{"issue": string(issue)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Interview, 0, len(records))
	for _, rec := range records {
		interview, err := toInterview(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, interview)
	}
	return out, nil
}

func (r *InterviewRepository) Update(ctx context.Context, i domain.Interview) (domain.Interview, error) {
	rec, err := r.app.FindRecordById(ColInterviews, string(i.ID))
	if err != nil {
		return domain.Interview{}, mapErr(err)
	}
	applyInterview(rec, i)
	if err := r.app.Save(rec); err != nil {
		return domain.Interview{}, mapErr(err)
	}
	return toInterview(rec)
}

func (r *InterviewRepository) AppendEvent(ctx context.Context, i domain.Interview, actions []domain.SendAction, at time.Time) (domain.InterviewEvent, error) {
	var event domain.InterviewEvent
	err := r.app.RunInTransaction(func(tx core.App) error {
		var last int
		if err := tx.DB().NewQuery("SELECT COALESCE(MAX(seq), 0) FROM " + ColInterviewEvents + " WHERE interview = {:interview}").
			Bind(dbx.Params{"interview": string(i.ID)}).Row(&last); err != nil {
			return err
		}
		collection, err := tx.FindCollectionByNameOrId(ColInterviewEvents)
		if err != nil {
			return err
		}
		rec := core.NewRecord(collection)
		rec.Set("project", string(i.ProjectID))
		rec.Set("interview", string(i.ID))
		rec.Set("seq", last+1)
		setDate(rec, "at", &at)
		rec.Set("actions", actions)
		if err := tx.Save(rec); err != nil {
			return err
		}
		event, err = toEvent(rec)
		return err
	})
	if err != nil {
		return domain.InterviewEvent{}, mapErr(err)
	}
	return event, nil
}

func (r *InterviewRepository) EventsAfter(ctx context.Context, id domain.InterviewID, seq int) ([]domain.InterviewEvent, error) {
	records, err := r.app.FindRecordsByFilter(ColInterviewEvents, "interview = {:interview} && seq > {:seq}", "seq", 0, 0,
		dbx.Params{"interview": string(id), "seq": seq})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.InterviewEvent, 0, len(records))
	for _, rec := range records {
		event, err := toEvent(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, nil
}

func toEvent(rec *core.Record) (domain.InterviewEvent, error) {
	var actions []domain.SendAction
	if raw := rec.GetString("actions"); raw != "" && raw != "null" {
		if err := json.Unmarshal([]byte(raw), &actions); err != nil {
			return domain.InterviewEvent{}, fmt.Errorf("interview event %s: read actions: %w", rec.Id, err)
		}
	}
	return domain.InterviewEvent{
		ID:          domain.EventID(rec.Id),
		InterviewID: domain.InterviewID(rec.GetString("interview")),
		Seq:         rec.GetInt("seq"),
		At:          rec.GetDateTime("at").Time(),
		Actions:     actions,
	}, nil
}
