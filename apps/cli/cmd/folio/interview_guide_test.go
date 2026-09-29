package main

import (
	"strings"
	"testing"

	"folio/cli/internal/client"
)

func question(id, status string) client.InterviewQuestion {
	return client.InterviewQuestion{ID: id, Title: "Title " + id, Status: status}
}

func guideText(lines []string) string {
	return strings.Join(lines, "\n")
}

func TestGuideAsksForRoundOneWhenThereAreNoQuestions(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{}))
	for _, want := range []string{"round 1", "1 to 3", "folio interview patch tk"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestGuideHandsTheLinkOverWhileQuestionsAreOpen(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Questions: []client.InterviewQuestion{question("q1", "answered"), question("q2", "reopened")}}))
	for _, want := range []string{"link", "end the turn", "folio interview pending tk"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
	if strings.Contains(text, "finish") {
		t.Errorf("an open question blocks finish, got %q", text)
	}
}

func TestGuideOffersFinishOnceEveryQuestionIsSettled(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Questions: []client.InterviewQuestion{question("q1", "answered"), question("q2", "deferred")}}))
	for _, want := range []string{"next round", "confirm", "folio interview finish tk"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestGuideAppliesPendingSendsInOnePatch(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Questions: []client.InterviewQuestion{question("q1", "open")}, Pending: pendingSome, LastSeq: 7}))
	for _, want := range []string{"one patch", `"handled":7`, "separate patches"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestGuideStopsWhenNothingWasSent(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Questions: []client.InterviewQuestion{question("q1", "open")}, Pending: pendingNone}))
	if !strings.Contains(text, "stop") || strings.Contains(text, "patch") {
		t.Errorf("nothing sent means stop, got %q", text)
	}
}

func TestGuideRoutesEveryDeferredQuestionAfterFinish(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Finished: true, Questions: []client.InterviewQuestion{question("q1", "answered"), question("q2", "deferred")}}))
	for _, want := range []string{"Title q2", "decision ticket", "Not yet specified", "doc only"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
	if strings.Contains(text, "Title q1") {
		t.Errorf("an answered question needs no routing, got %q", text)
	}
}

func TestGuideEndsQuietlyAfterFinishWithNothingDeferred(t *testing.T) {
	text := guideText(interviewGuide("tk", interviewState{Finished: true, Questions: []client.InterviewQuestion{question("q1", "answered")}}))
	if !strings.Contains(text, "resolved") || strings.Contains(text, "decision ticket") {
		t.Errorf("got %q", text)
	}
}
