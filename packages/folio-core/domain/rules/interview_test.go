package rules_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

var at = time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)

func question(id string, round int) domain.Question {
	return domain.Question{
		ID: id, Round: round, Title: "Tree or graph", Status: domain.QuestionOpen,
		Options: []domain.Option{{Key: "a", Text: "Tree"}, {Key: "b", Text: "Graph"}},
		Rec:     domain.Recommendation{Option: "b", Why: "The tree hides blockers."},
		Deps:    []string{}, Thread: []domain.Message{},
	}
}

func state(qs ...domain.Question) domain.InterviewState {
	return domain.InterviewState{Terms: []domain.Term{}, Questions: qs}
}

func invalid(t *testing.T, err error, why string) {
	t.Helper()
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("%s: want a validation error, got %v", why, err)
	}
}

func TestValidateInterview(t *testing.T) {
	t.Run("accepts a well formed state", func(t *testing.T) {
		free := question("q2", 1)
		free.Options = []domain.Option{}
		free.Rec = domain.Recommendation{Text: "Name it interview", Why: "Plain word."}
		if err := rules.ValidateInterview(state(question("q1", 1), free)); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	cases := map[string]func(q *domain.Question){
		"a missing title":  func(q *domain.Question) { q.Title = "" },
		"a title past 200": func(q *domain.Question) { q.Title = strings.Repeat("x", 201) },
		"a body past 2000": func(q *domain.Question) { q.Body = strings.Repeat("x", 2001) },
		"a why past 2000":  func(q *domain.Question) { q.Rec.Why = strings.Repeat("x", 2001) },
		"one option":       func(q *domain.Question) { q.Options = q.Options[:1] },
		"five options": func(q *domain.Question) {
			q.Options = append(q.Options, domain.Option{Key: "c", Text: "c"}, domain.Option{Key: "d", Text: "d"}, domain.Option{Key: "e", Text: "e"})
		},
		"an option key outside a to d":  func(q *domain.Question) { q.Options[1].Key = "z" },
		"a repeated option key":         func(q *domain.Question) { q.Options[1].Key = "a" },
		"a recommendation off the list": func(q *domain.Question) { q.Rec.Option = "c" },
		"a text rec beside options":     func(q *domain.Question) { q.Rec = domain.Recommendation{Text: "x", Why: "y"} },
		"an unknown status":             func(q *domain.Question) { q.Status = "maybe" },
		"answered without an answer":    func(q *domain.Question) { q.Status = domain.QuestionAnswered },
		"an answer on a missing option": func(q *domain.Question) {
			q.Status = domain.QuestionAnswered
			q.Answer = &domain.Answer{Kind: domain.AnswerOption, Option: "c"}
		},
		"accept that is not the rec": func(q *domain.Question) {
			q.Status = domain.QuestionAnswered
			q.Answer = &domain.Answer{Kind: domain.AnswerAccept, Option: "a"}
		},
		"an empty text answer": func(q *domain.Question) {
			q.Status = domain.QuestionAnswered
			q.Answer = &domain.Answer{Kind: domain.AnswerText}
		},
		"a dep that does not exist": func(q *domain.Question) { q.Deps = []string{"q9"} },
		"a round below 1":           func(q *domain.Question) { q.Round = 0 },
		"a message past 4000": func(q *domain.Question) {
			q.Thread = []domain.Message{{Who: domain.SpeakerUser, Text: strings.Repeat("x", 4001)}}
		},
		"an unknown speaker": func(q *domain.Question) { q.Thread = []domain.Message{{Who: "bot", Text: "hi"}} },
		"an explore row off the list": func(q *domain.Question) {
			q.Explore = &domain.Explore{Rows: []domain.ExploreRow{{Option: "c", Pros: []string{"x", "y"}, Cons: []string{"x", "y"}}}}
		},
		"an explore row with one pro": func(q *domain.Question) {
			q.Explore = &domain.Explore{Rows: []domain.ExploreRow{{Option: "a", Pros: []string{"x"}, Cons: []string{"x", "y"}}}}
		},
		"an explore item past 200": func(q *domain.Question) {
			q.Explore = &domain.Explore{Rows: []domain.ExploreRow{{Option: "a", Pros: []string{strings.Repeat("x", 201), "y"}, Cons: []string{"x", "y"}}}}
		},
	}
	for name, spoil := range cases {
		t.Run("refuses "+name, func(t *testing.T) {
			q := question("q1", 1)
			spoil(&q)
			invalid(t, rules.ValidateInterview(state(q)), name)
		})
	}

	t.Run("refuses a repeated question id", func(t *testing.T) {
		invalid(t, rules.ValidateInterview(state(question("q1", 1), question("q1", 2))), "repeated id")
	})

	t.Run("refuses a dependency cycle", func(t *testing.T) {
		a, b := question("q1", 1), question("q2", 2)
		a.Deps, b.Deps = []string{"q2"}, []string{"q1"}
		invalid(t, rules.ValidateInterview(state(a, b)), "cycle")
	})

	t.Run("refuses a fourth question in a round", func(t *testing.T) {
		invalid(t, rules.ValidateInterview(state(question("q1", 1), question("q2", 1), question("q3", 1), question("q4", 1))), "round cap")
	})

	t.Run("refuses a term definition past 500", func(t *testing.T) {
		s := state(question("q1", 1))
		s.Terms = []domain.Term{{Term: "interview", Def: strings.Repeat("x", 501), Avoid: []string{}}}
		invalid(t, rules.ValidateInterview(s), "term def")
	})
}

func TestApplyInterviewPatch(t *testing.T) {
	base := state(question("q1", 1))

	t.Run("adds a round and returns handled", func(t *testing.T) {
		patch := `{"questions":[{"id":"q2","round":2,"deps":["q1"],"title":"Where does it live","options":[{"k":"a","text":"One record"},{"k":"b","text":"Many"}],"rec":{"option":"a","why":"Atomic."}}],"agent":{"handled":3}}`
		next, handled, err := rules.ApplyInterviewPatch(base, []byte(patch), at)
		if err != nil {
			t.Fatal(err)
		}
		if len(next.Questions) != 2 || next.Questions[1].Status != domain.QuestionOpen {
			t.Fatalf("questions = %+v", next.Questions)
		}
		if handled == nil || *handled != 3 {
			t.Fatalf("handled = %v, want 3", handled)
		}
	})

	t.Run("merges a known question one level and deletes with null", func(t *testing.T) {
		answered := base
		answered.Questions = []domain.Question{question("q1", 1)}
		answered.Questions[0].Status = domain.QuestionAnswered
		answered.Questions[0].Answer = &domain.Answer{Kind: domain.AnswerAccept, Option: "b"}

		next, _, err := rules.ApplyInterviewPatch(answered, []byte(`{"questions":[{"id":"q1","status":"reopened","answer":null}]}`), at)
		if err != nil {
			t.Fatal(err)
		}
		got := next.Questions[0]
		if got.Status != domain.QuestionReopened || got.Answer != nil || got.Title != "Tree or graph" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("appends to the thread and stamps its time", func(t *testing.T) {
		next, _, err := rules.ApplyInterviewPatch(base, []byte(`{"questions":[{"id":"q1","thread":[{"who":"agent","text":"Because edges show blockers."}]}]}`), at)
		if err != nil {
			t.Fatal(err)
		}
		thread := next.Questions[0].Thread
		if len(thread) != 1 || thread[0].Who != domain.SpeakerAgent || !thread[0].At.Equal(at) {
			t.Fatalf("thread = %+v", thread)
		}
	})

	t.Run("replaces a known term whole", func(t *testing.T) {
		withTerm := base
		withTerm.Terms = []domain.Term{{Term: "interview", Def: "old", Avoid: []string{"grill"}}}
		next, _, err := rules.ApplyInterviewPatch(withTerm, []byte(`{"terms":[{"term":"interview","def":"new"}]}`), at)
		if err != nil {
			t.Fatal(err)
		}
		if len(next.Terms) != 1 || next.Terms[0].Def != "new" || len(next.Terms[0].Avoid) != 0 {
			t.Fatalf("terms = %+v", next.Terms)
		}
	})

	for name, patch := range map[string]string{
		"an unknown top level key":          `{"visual":{}}`,
		"an unknown question key":           `{"questions":[{"id":"q1","colour":"red"}]}`,
		"an unknown question without title": `{"questions":[{"id":"q7","round":2}]}`,
		"a patch that breaks the shape":     `{"questions":[{"id":"q1","options":[{"k":"a","text":"only one"}]}]}`,
		"an agent key other than handled":   `{"agent":{"status":"working"}}`,
		"an unknown key inside rec":         `{"questions":[{"id":"q1","rec":{"option":"b","why":"x","colour":"red"}}]}`,
		"an unknown key inside an answer":   `{"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b","by":"me"}}]}`,
		"an unknown key inside a term":      `{"terms":[{"term":"interview","def":"x","colour":"red"}]}`,
		"not json":                          `{"questions":`,
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			_, _, err := rules.ApplyInterviewPatch(base, []byte(patch), at)
			invalid(t, err, name)
		})
	}

	t.Run("leaves the original state untouched", func(t *testing.T) {
		if _, _, err := rules.ApplyInterviewPatch(base, []byte(`{"questions":[{"id":"q1","title":"Renamed"}]}`), at); err != nil {
			t.Fatal(err)
		}
		if base.Questions[0].Title != "Tree or graph" {
			t.Fatal("the patch mutated its input")
		}
	})
}

func TestValidateSend(t *testing.T) {
	open := question("q1", 1)
	free := question("q2", 1)
	free.Options = []domain.Option{}
	free.Rec = domain.Recommendation{Text: "Say it", Why: "why"}
	s := state(open, free)

	ok := [][]domain.SendAction{
		{{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerAccept, Option: "b"}},
		{{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerOption, Option: "a"}},
		{{Type: domain.SendAnswer, Q: "q2", Kind: domain.AnswerText, Text: "interview"}},
		{{Type: domain.SendThread, Q: "q1", Text: "why not a tree?"}, {Type: domain.SendExplore, Q: "q1"}},
		{{Type: domain.SendDefer, Q: "q1"}},
		{{Type: domain.SendFinish}},
	}
	for _, actions := range ok {
		if err := rules.ValidateSend(s, actions); err != nil {
			t.Errorf("%+v: want nil, got %v", actions, err)
		}
	}

	for name, actions := range map[string][]domain.SendAction{
		"no actions":                 {},
		"an unknown type":            {{Type: "visualize"}},
		"a question that is missing": {{Type: domain.SendDefer, Q: "q9"}},
		"an option that is missing":  {{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerOption, Option: "c"}},
		"accept off the rec":         {{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerAccept, Option: "a"}},
		"an empty thread message":    {{Type: domain.SendThread, Q: "q1"}},
		"explore on a free question": {{Type: domain.SendExplore, Q: "q2"}},
		"an unknown answer kind":     {{Type: domain.SendAnswer, Q: "q1", Kind: "maybe"}},
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			invalid(t, rules.ValidateSend(s, actions), name)
		})
	}
}

func TestValidateSendDeferAndReopen(t *testing.T) {
	answered := question("q1", 1)
	answered.Status = domain.QuestionAnswered
	answered.Answer = &domain.Answer{Kind: domain.AnswerAccept, Option: "b"}
	s := state(answered, question("q2", 1))

	if err := rules.ValidateSend(s, []domain.SendAction{{Type: domain.SendReopen, Q: "q1"}, {Type: domain.SendDefer, Q: "q2"}}); err != nil {
		t.Fatalf("reopen an answered question and defer an open one: %v", err)
	}
	invalid(t, rules.ValidateSend(s, []domain.SendAction{{Type: domain.SendDefer, Q: "q1"}}), "defer an answered question")
	invalid(t, rules.ValidateSend(s, []domain.SendAction{{Type: domain.SendReopen, Q: "q2"}}), "reopen an open question")
}

func TestCheckFinishable(t *testing.T) {
	deferred := question("q2", 1)
	deferred.Status = domain.QuestionDeferred
	answered := question("q1", 1)
	answered.Status = domain.QuestionAnswered
	answered.Answer = &domain.Answer{Kind: domain.AnswerAccept, Option: "b"}

	if err := rules.CheckFinishable(state(answered, deferred)); err != nil {
		t.Fatalf("every question settled: %v", err)
	}
	reopened := question("q3", 2)
	reopened.Status = domain.QuestionReopened
	invalid(t, rules.CheckFinishable(state(answered, reopened)), "a reopened question is open")
	invalid(t, rules.CheckFinishable(state()), "an interview with no questions has nothing to finish")
}

func TestCheckHandled(t *testing.T) {
	if err := rules.CheckHandled(2, 4, 5); err != nil {
		t.Fatalf("forward within the Sends: %v", err)
	}
	if err := rules.CheckHandled(4, 4, 4); err != nil {
		t.Fatalf("the same seq again is a no-op: %v", err)
	}
	invalid(t, rules.CheckHandled(4, 3, 5), "backwards")
	invalid(t, rules.CheckHandled(2, 6, 5), "past the last Send")
}
