package rules

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"folio/folio-core/domain"
)

const (
	MessageMaxLen     = 4000
	TermDefMaxLen     = 500
	ExploreItemMaxLen = 200
	QuestionIDMaxLen  = 20
	QuestionsPerRound = 3
	QuestionsMax      = 200
	OptionsMin        = 2
	OptionsMax        = 4
	ExploreItemsMin   = 2
	ExploreItemsMax   = 4
	optionKeys        = "abcd"
)

func ValidateInterview(s domain.InterviewState) error {
	if len(s.Questions) > QuestionsMax {
		return domain.Invalid("questions", "at most "+strconv.Itoa(QuestionsMax)+" per interview")
	}
	for _, term := range s.Terms {
		if err := validateTerm(term); err != nil {
			return err
		}
	}

	byID := make(map[string]domain.Question, len(s.Questions))
	perRound := map[int]int{}
	for _, q := range s.Questions {
		if q.ID == "" || len(q.ID) > QuestionIDMaxLen {
			return domain.Invalid("question.id", "is required and at most "+strconv.Itoa(QuestionIDMaxLen)+" characters")
		}
		if _, seen := byID[q.ID]; seen {
			return domain.Invalid("question.id", q.ID+" is used twice")
		}
		byID[q.ID] = q
		if q.Round < 1 {
			return domain.Invalid(q.ID+".round", "must be 1 or greater")
		}
		perRound[q.Round]++
		if perRound[q.Round] > QuestionsPerRound {
			return domain.Invalid(q.ID+".round", "a round holds at most "+strconv.Itoa(QuestionsPerRound)+" questions")
		}
	}

	for _, q := range s.Questions {
		if err := validateQuestion(q, byID); err != nil {
			return err
		}
	}
	return checkNoDepCycle(s.Questions)
}

func validateTerm(t domain.Term) error {
	if err := required("term", t.Term, NameMaxLen); err != nil {
		return err
	}
	if err := required("term.def", t.Def, TermDefMaxLen); err != nil {
		return err
	}
	for _, word := range t.Avoid {
		if err := required("term.avoid", word, NameMaxLen); err != nil {
			return err
		}
	}
	return nil
}

func validateQuestion(q domain.Question, byID map[string]domain.Question) error {
	field := func(name string) string { return q.ID + "." + name }

	if err := required(field("title"), q.Title, TitleMaxLen); err != nil {
		return err
	}
	if err := optional(field("body"), q.Body, DescrMaxLen); err != nil {
		return err
	}
	for _, dep := range q.Deps {
		if _, ok := byID[dep]; !ok {
			return domain.Invalid(field("deps"), dep+" is not a question in this interview")
		}
	}

	keys, err := optionSet(q)
	if err != nil {
		return err
	}
	if err := validateRec(q, keys); err != nil {
		return err
	}

	switch q.Status {
	case domain.QuestionOpen, domain.QuestionDeferred, domain.QuestionReopened:
		if q.Answer != nil {
			return domain.Invalid(field("answer"), "only an answered question carries an answer")
		}
	case domain.QuestionAnswered:
		if q.Answer == nil {
			return domain.Invalid(field("answer"), "an answered question needs its answer")
		}
		if err := validateAnswer(q, keys, *q.Answer); err != nil {
			return err
		}
	default:
		return domain.Invalid(field("status"), "must be one of open, answered, deferred, reopened")
	}

	if q.Explore != nil {
		if err := validateExplore(q, keys, *q.Explore); err != nil {
			return err
		}
	}
	for _, m := range q.Thread {
		if err := validateMessage(field("thread"), m); err != nil {
			return err
		}
	}
	return nil
}

func optionSet(q domain.Question) (map[string]bool, error) {
	field := q.ID + ".options"
	if len(q.Options) == 0 {
		return map[string]bool{}, nil
	}
	if len(q.Options) < OptionsMin || len(q.Options) > OptionsMax {
		return nil, domain.Invalid(field, "must be empty or hold 2 to 4 options")
	}
	keys := make(map[string]bool, len(q.Options))
	for _, o := range q.Options {
		if len(o.Key) != 1 || !bytes.ContainsAny([]byte(optionKeys), o.Key) {
			return nil, domain.Invalid(field, "keys are a single letter from a to d")
		}
		if keys[o.Key] {
			return nil, domain.Invalid(field, o.Key+" is used twice")
		}
		keys[o.Key] = true
		if err := required(field, o.Text, TitleMaxLen); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func validateRec(q domain.Question, keys map[string]bool) error {
	field := q.ID + ".rec"
	if err := required(field+".why", q.Rec.Why, DescrMaxLen); err != nil {
		return err
	}
	if len(keys) > 0 {
		if q.Rec.Text != "" || !keys[q.Rec.Option] {
			return domain.Invalid(field, "must name one of the options")
		}
		return nil
	}
	if q.Rec.Option != "" {
		return domain.Invalid(field, "a question without options recommends text, not an option")
	}
	return required(field+".text", q.Rec.Text, DescrMaxLen)
}

func validateAnswer(q domain.Question, keys map[string]bool, a domain.Answer) error {
	field := q.ID + ".answer"
	switch a.Kind {
	case domain.AnswerAccept:
		if len(keys) == 0 || a.Option != q.Rec.Option {
			return domain.Invalid(field, "accept takes the recommended option")
		}
	case domain.AnswerOption:
		if !keys[a.Option] {
			return domain.Invalid(field, "must name one of the options")
		}
	case domain.AnswerText:
		return required(field+".text", a.Text, MessageMaxLen)
	default:
		return domain.Invalid(field+".kind", "must be one of accept, option, text")
	}
	return nil
}

func validateExplore(q domain.Question, keys map[string]bool, e domain.Explore) error {
	field := q.ID + ".explore"
	if len(keys) == 0 {
		return domain.Invalid(field, "only a question with options can be explored")
	}
	for _, row := range e.Rows {
		if !keys[row.Option] {
			return domain.Invalid(field, row.Option+" is not one of the options")
		}
		for name, items := range map[string][]string{"pros": row.Pros, "cons": row.Cons} {
			if len(items) < ExploreItemsMin || len(items) > ExploreItemsMax {
				return domain.Invalid(field+"."+name, "each option takes 2 to 4")
			}
			for _, item := range items {
				if err := required(field+"."+name, item, ExploreItemMaxLen); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateMessage(field string, m domain.Message) error {
	if m.Who != domain.SpeakerUser && m.Who != domain.SpeakerAgent {
		return domain.Invalid(field+".who", "must be user or agent")
	}
	return required(field+".text", m.Text, MessageMaxLen)
}

func checkNoDepCycle(questions []domain.Question) error {
	deps := make(map[string][]string, len(questions))
	for _, q := range questions {
		deps[q.ID] = q.Deps
	}
	const (
		unseen = iota
		walking
		done
	)
	state := make(map[string]int, len(questions))
	var walk func(id string) bool
	walk = func(id string) bool {
		switch state[id] {
		case walking:
			return true
		case done:
			return false
		}
		state[id] = walking
		for _, dep := range deps[id] {
			if walk(dep) {
				return true
			}
		}
		state[id] = done
		return false
	}
	for _, q := range questions {
		if walk(q.ID) {
			return domain.Invalid(q.ID+".deps", "introduces a dependency cycle")
		}
	}
	return nil
}

var (
	patchKeys    = map[string]bool{"note": true, "terms": true, "questions": true, "agent": true}
	questionKeys = map[string]bool{
		"id": true, "round": true, "deps": true, "title": true, "body": true, "options": true, "rec": true,
		"status": true, "durable": true, "updated": true, "answer": true, "explore": true, "thread": true,
	}
)

func ApplyInterviewPatch(s domain.InterviewState, raw []byte, now time.Time) (domain.InterviewState, *int, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return domain.InterviewState{}, nil, domain.Invalid("patch", "is not a JSON object")
	}
	for key := range top {
		if !patchKeys[key] {
			return domain.InterviewState{}, nil, domain.Invalid("patch", key+" is not a patch key; use note, terms, questions or agent")
		}
	}

	next, err := cloneState(s)
	if err != nil {
		return domain.InterviewState{}, nil, err
	}

	if value, ok := top["note"]; ok {
		next.Note = ""
		if !isNull(value) {
			if err := strict(value, &next.Note); err != nil {
				return domain.InterviewState{}, nil, domain.Invalid("note", "must be a string or null")
			}
		}
	}

	if value, ok := top["terms"]; ok {
		if err := patchTerms(&next, value); err != nil {
			return domain.InterviewState{}, nil, err
		}
	}

	if value, ok := top["questions"]; ok {
		if err := patchQuestions(&next, value, now); err != nil {
			return domain.InterviewState{}, nil, err
		}
	}

	var handled *int
	if value, ok := top["agent"]; ok {
		var agent struct {
			Handled *int `json:"handled"`
		}
		if err := strict(value, &agent); err != nil || agent.Handled == nil {
			return domain.InterviewState{}, nil, domain.Invalid("agent", "takes only handled, the seq of the last Send applied")
		}
		handled = agent.Handled
	}

	if err := ValidateInterview(next); err != nil {
		return domain.InterviewState{}, nil, err
	}
	return next, handled, nil
}

func patchTerms(s *domain.InterviewState, value json.RawMessage) error {
	var terms []json.RawMessage
	if err := strict(value, &terms); err != nil {
		return domain.Invalid("terms", "must be a list of terms")
	}
	for _, raw := range terms {
		var term domain.Term
		if err := strict(raw, &term); err != nil {
			return domain.Invalid("terms", "a term takes term, def and avoid")
		}
		if term.Avoid == nil {
			term.Avoid = []string{}
		}
		replaced := false
		for i := range s.Terms {
			if s.Terms[i].Term == term.Term {
				s.Terms[i], replaced = term, true
			}
		}
		if !replaced {
			s.Terms = append(s.Terms, term)
		}
	}
	return nil
}

func patchQuestions(s *domain.InterviewState, value json.RawMessage, now time.Time) error {
	var patches []json.RawMessage
	if err := strict(value, &patches); err != nil {
		return domain.Invalid("questions", "must be a list of question patches")
	}
	for _, raw := range patches {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return domain.Invalid("questions", "each patch is an object")
		}
		var id string
		if err := strict(fields["id"], &id); err != nil || id == "" {
			return domain.Invalid("question.id", "every question patch names its id")
		}

		index := -1
		for i := range s.Questions {
			if s.Questions[i].ID == id {
				index = i
			}
		}
		if index < 0 {
			if _, ok := fields["title"]; !ok {
				return domain.Invalid(id, "is not a question yet; a new question needs a title")
			}
			s.Questions = append(s.Questions, domain.Question{
				ID: id, Status: domain.QuestionOpen, Deps: []string{}, Options: []domain.Option{}, Thread: []domain.Message{},
			})
			index = len(s.Questions) - 1
		}
		if err := patchQuestion(&s.Questions[index], fields, now); err != nil {
			return err
		}
	}
	return nil
}

func patchQuestion(q *domain.Question, fields map[string]json.RawMessage, now time.Time) error {
	for key, value := range fields {
		if !questionKeys[key] {
			return domain.Invalid(q.ID, key+" is not a question field")
		}
		if key == "id" {
			continue
		}
		field := q.ID + "." + key
		if isNull(value) {
			if key == "thread" || key == "title" || key == "round" || key == "rec" || key == "status" {
				return domain.Invalid(field, "cannot be deleted")
			}
			clearQuestionField(q, key)
			continue
		}
		var err error
		switch key {
		case "round":
			err = strict(value, &q.Round)
		case "deps":
			err = strict(value, &q.Deps)
		case "title":
			err = strict(value, &q.Title)
		case "body":
			err = strict(value, &q.Body)
		case "options":
			err = strict(value, &q.Options)
		case "rec":
			var rec domain.Recommendation
			err = strict(value, &rec)
			q.Rec = rec
		case "status":
			err = strict(value, &q.Status)
		case "durable":
			err = strict(value, &q.Durable)
		case "updated":
			err = strict(value, &q.Updated)
		case "answer":
			var answer domain.Answer
			err = strict(value, &answer)
			q.Answer = &answer
		case "explore":
			var explore domain.Explore
			err = strict(value, &explore)
			if explore.At.IsZero() {
				explore.At = now
			}
			q.Explore = &explore
		case "thread":
			var messages []domain.Message
			err = strict(value, &messages)
			for i := range messages {
				if messages[i].At.IsZero() {
					messages[i].At = now
				}
			}
			q.Thread = append(q.Thread, messages...)
		}
		if err != nil {
			return domain.Invalid(field, "has the wrong shape: "+err.Error())
		}
	}
	if q.Deps == nil {
		q.Deps = []string{}
	}
	if q.Options == nil {
		q.Options = []domain.Option{}
	}
	return nil
}

func clearQuestionField(q *domain.Question, key string) {
	switch key {
	case "deps":
		q.Deps = []string{}
	case "body":
		q.Body = ""
	case "options":
		q.Options = []domain.Option{}
	case "durable":
		q.Durable = false
	case "updated":
		q.Updated = false
	case "answer":
		q.Answer = nil
	case "explore":
		q.Explore = nil
	}
}

func ValidateSend(s domain.InterviewState, actions []domain.SendAction) error {
	if len(actions) == 0 {
		return domain.Invalid("actions", "a Send carries at least one action")
	}
	byID := make(map[string]domain.Question, len(s.Questions))
	for _, q := range s.Questions {
		byID[q.ID] = q
	}
	for _, a := range actions {
		if a.Type == domain.SendFinish {
			continue
		}
		q, ok := byID[a.Q]
		switch a.Type {
		case domain.SendAnswer, domain.SendThread, domain.SendDefer, domain.SendReopen, domain.SendExplore:
			if !ok {
				return domain.Invalid("actions.q", a.Q+" is not a question in this interview")
			}
		default:
			return domain.Invalid("actions.type", "must be one of answer, thread, defer, reopen, explore, finish")
		}
		keys, err := optionSet(q)
		if err != nil {
			return err
		}
		switch a.Type {
		case domain.SendAnswer:
			if err := validateAnswer(q, keys, domain.Answer{Kind: a.Kind, Option: a.Option, Text: a.Text}); err != nil {
				return err
			}
		case domain.SendThread:
			if err := required(a.Q+".thread", a.Text, MessageMaxLen); err != nil {
				return err
			}
		case domain.SendExplore:
			if len(keys) == 0 {
				return domain.Invalid(a.Q, "only a question with options can be explored")
			}
		case domain.SendDefer:
			if !q.Status.IsOpen() {
				return domain.Invalid(a.Q, "only an open question can be deferred")
			}
		case domain.SendReopen:
			if q.Status.IsOpen() {
				return domain.Invalid(a.Q, "is already open")
			}
		}
	}
	return nil
}

func cloneState(s domain.InterviewState) (domain.InterviewState, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return domain.InterviewState{}, err
	}
	var out domain.InterviewState
	if err := json.Unmarshal(raw, &out); err != nil {
		return domain.InterviewState{}, err
	}
	if out.Terms == nil {
		out.Terms = []domain.Term{}
	}
	if out.Questions == nil {
		out.Questions = []domain.Question{}
	}
	return out, nil
}

func strict(raw json.RawMessage, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(into)
}

func isNull(raw json.RawMessage) bool {
	return string(bytes.TrimSpace(raw)) == "null"
}

func CheckFinishable(s domain.InterviewState) error {
	if len(s.Questions) == 0 {
		return domain.Invalid("finish", "the interview has no questions yet")
	}
	for _, q := range s.Questions {
		if q.Status.IsOpen() {
			return domain.Invalid("finish", q.ID+" is still open; answer or defer it first")
		}
	}
	return nil
}

func CheckHandled(current, next, lastSeq int) error {
	if next < current {
		return domain.Invalid("agent.handled", "cannot move back from "+strconv.Itoa(current))
	}
	if next > lastSeq {
		return domain.Invalid("agent.handled", "the last Send is "+strconv.Itoa(lastSeq))
	}
	return nil
}

func CheckInterviewable(ticket domain.Issue) error {
	if ticket.Kind != domain.IssueTicket || ticket.Wayfinder != domain.WayfinderGrilling {
		return domain.Invalid("ticket", "an interview runs on a grilling ticket")
	}
	if ticket.Status.IsTerminal() {
		return domain.Invalid("ticket", "is closed; reopen it to ask the question again")
	}
	return nil
}

type PatchSummary struct {
	Round    int
	Added    int
	Answered int
}

func SummarisePatch(before, after domain.InterviewState) PatchSummary {
	known := make(map[string]domain.QuestionStatus, len(before.Questions))
	for _, q := range before.Questions {
		known[q.ID] = q.Status
	}
	var out PatchSummary
	for _, q := range after.Questions {
		if q.Round > out.Round {
			out.Round = q.Round
		}
		status, existed := known[q.ID]
		if !existed {
			out.Added++
		}
		if q.Status == domain.QuestionAnswered && status != domain.QuestionAnswered {
			out.Answered++
		}
	}
	return out
}
