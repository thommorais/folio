package domain

import "time"

type QuestionStatus string

const (
	QuestionOpen     QuestionStatus = "open"
	QuestionAnswered QuestionStatus = "answered"
	QuestionDeferred QuestionStatus = "deferred"
	QuestionReopened QuestionStatus = "reopened"
)

func (s QuestionStatus) IsOpen() bool { return s == QuestionOpen || s == QuestionReopened }

type AnswerKind string

const (
	AnswerAccept AnswerKind = "accept"
	AnswerOption AnswerKind = "option"
	AnswerText   AnswerKind = "text"
)

type AgentStatus string

const (
	AgentWaiting AgentStatus = "waiting"
	AgentWorking AgentStatus = "working"
)

type Speaker string

const (
	SpeakerUser  Speaker = "user"
	SpeakerAgent Speaker = "agent"
)

type Option struct {
	Key  string `json:"k"`
	Text string `json:"text"`
}

type Recommendation struct {
	Option string `json:"option,omitempty"`
	Text   string `json:"text,omitempty"`
	Why    string `json:"why"`
}

type Answer struct {
	Kind   AnswerKind `json:"kind"`
	Option string     `json:"option,omitempty"`
	Text   string     `json:"text,omitempty"`
}

type ExploreRow struct {
	Option string   `json:"option"`
	Pros   []string `json:"pros"`
	Cons   []string `json:"cons"`
}

type Explore struct {
	At   time.Time    `json:"at"`
	Rows []ExploreRow `json:"rows"`
}

type Message struct {
	Who  Speaker   `json:"who"`
	Text string    `json:"text"`
	At   time.Time `json:"at"`
}

type Question struct {
	ID      string         `json:"id"`
	Round   int            `json:"round"`
	Deps    []string       `json:"deps"`
	Title   string         `json:"title"`
	Body    string         `json:"body,omitempty"`
	Options []Option       `json:"options"`
	Rec     Recommendation `json:"rec"`
	Status  QuestionStatus `json:"status"`
	Durable bool           `json:"durable"`
	Updated bool           `json:"updated"`
	Answer  *Answer        `json:"answer,omitempty"`
	Explore *Explore       `json:"explore,omitempty"`
	Thread  []Message      `json:"thread"`
}

type Term struct {
	Term  string   `json:"term"`
	Def   string   `json:"def"`
	Avoid []string `json:"avoid"`
}

type InterviewState struct {
	Note      string     `json:"note,omitempty"`
	Terms     []Term     `json:"terms"`
	Questions []Question `json:"questions"`
}

type Interview struct {
	ID          InterviewID
	ProjectID   ProjectID
	IssueID     IssueID
	Topic       string
	State       InterviewState
	AgentStatus AgentStatus
	AgentSince  time.Time
	Handled     int
	FinishedAt  *time.Time
	CreatedBy   UserID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (i Interview) IsFinished() bool { return i.FinishedAt != nil }

type SendActionType string

const (
	SendAnswer  SendActionType = "answer"
	SendThread  SendActionType = "thread"
	SendDefer   SendActionType = "defer"
	SendReopen  SendActionType = "reopen"
	SendExplore SendActionType = "explore"
	SendFinish  SendActionType = "finish"
)

type SendAction struct {
	Type   SendActionType `json:"type"`
	Q      string         `json:"q,omitempty"`
	Kind   AnswerKind     `json:"kind,omitempty"`
	Option string         `json:"option,omitempty"`
	Text   string         `json:"text,omitempty"`
}

type InterviewEvent struct {
	ID          EventID
	InterviewID InterviewID
	Seq         int
	At          time.Time
	Actions     []SendAction
}
