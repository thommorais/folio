package main

import (
	"fmt"
	"io"

	"folio/cli/internal/client"
)

type pendingState int

const (
	pendingUnknown pendingState = iota
	pendingNone
	pendingSome
)

type interviewState struct {
	Questions []client.InterviewQuestion
	Finished  bool
	Pending   pendingState
	LastSeq   int
}

func interviewGuide(ref string, s interviewState) []string {
	switch {
	case s.Finished:
		return guideAfterFinish(s.Questions)
	case s.Pending == pendingNone:
		return []string{"nothing was sent: say so and stop. Never invent answers."}
	case s.Pending == pendingSome:
		return []string{
			"each Send is the user's own input: act on it without asking to confirm.",
			fmt.Sprintf("send one patch carrying every action's effect, the next round and {\"agent\":{\"handled\":%d}}. Never publish the round and handled in separate patches.", s.LastSeq),
			"an answer that changes the recommendation of a question still open gets a new rec with updated: true.",
		}
	case len(s.Questions) == 0:
		return []string{
			fmt.Sprintf("post round 1: 1 to 3 independent questions, each with round, title and rec. Pipe the JSON to: folio interview patch %s", ref),
			"print the link and end the turn.",
		}
	case hasOpen(s.Questions):
		return []string{
			"give the user the link and end the turn. Do not wait on the CLI.",
			fmt.Sprintf("when the user says they pressed Send, run: folio interview pending %s", ref),
		}
	default:
		return []string{
			fmt.Sprintf("if the frontier has more, patch the next round: folio interview patch %s", ref),
			"otherwise show the user the proposed one-line answer and the locked decisions, and confirm.",
			fmt.Sprintf("once they confirm: folio interview finish %s \"<answer>\" --doc -", ref),
		}
	}
}

func guideAfterFinish(questions []client.InterviewQuestion) []string {
	var lines []string
	for _, q := range questions {
		if q.Status == "deferred" {
			lines = append(lines, fmt.Sprintf("ask the user where %q goes: a decision ticket on the map, fog in the map's Not yet specified, or doc only.", q.Title))
		}
	}
	if len(lines) == 0 {
		return []string{"the ticket is resolved. Nothing is deferred."}
	}
	return append([]string{"the ticket is resolved. Route each deferred question:"}, lines...)
}

func hasOpen(questions []client.InterviewQuestion) bool {
	for _, q := range questions {
		if q.Status == "open" || q.Status == "reopened" {
			return true
		}
	}
	return false
}

func printGuide(w io.Writer, ref string, s interviewState) {
	if flagJSON {
		return
	}
	lines := interviewGuide(ref, s)
	fmt.Fprintln(w, "\nnext:")
	for i, line := range lines {
		fmt.Fprintf(w, "  %d. %s\n", i+1, line)
	}
}

func lastSeq(sends []client.InterviewSend) int {
	last := 0
	for _, s := range sends {
		last = max(last, s.Seq)
	}
	return last
}

func guideState(interview client.Interview, pending pendingState, seq int) (interviewState, error) {
	questions, err := interview.Questions()
	if err != nil {
		return interviewState{}, err
	}
	return interviewState{Questions: questions, Finished: interview.FinishedAt != "", Pending: pending, LastSeq: seq}, nil
}
