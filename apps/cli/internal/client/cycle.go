package client

import "net/http"

type Cycle struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	TicketID   string `json:"issue_id"`
	Ordinal    int    `json:"ordinal"`
	Phase      string `json:"phase"`
	Resolution string `json:"resolution,omitempty"`
	MapID      string `json:"map_id,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ClosedAt   string `json:"closed_at,omitempty"`
}

type CycleInput struct {
	Phase      *string `json:"phase,omitempty"`
	Resolution *string `json:"resolution,omitempty"`
	MapID      *string `json:"map_id,omitempty"`
}

func (c *Client) ListCycles(ticket string) ([]Cycle, error) {
	var body struct {
		Cycles []Cycle `json:"cycles"`
	}
	if err := c.do(http.MethodGet, "/api/folio/issues/"+ticket+"/cycles", nil, &body); err != nil {
		return nil, err
	}
	return body.Cycles, nil
}

func (c *Client) OpenCycle(ticket string) (Cycle, error) {
	var cycle Cycle
	err := c.do(http.MethodPost, "/api/folio/issues/"+ticket+"/cycles", struct{}{}, &cycle)
	return cycle, err
}

func (c *Client) UpdateCurrentCycle(ticket string, in CycleInput) (Cycle, error) {
	var cycle Cycle
	err := c.do(http.MethodPatch, "/api/folio/issues/"+ticket+"/cycles/current", in, &cycle)
	return cycle, err
}

func (c *Client) NextPhase(ticket string) (Cycle, error) {
	var cycle Cycle
	err := c.do(http.MethodPost, "/api/folio/issues/"+ticket+"/cycles/current/next", struct{}{}, &cycle)
	return cycle, err
}

func (c *Client) OpenCycleWithMap(project, ticket, title string) (Cycle, Ticket, error) {
	if _, err := c.OpenCycle(ticket); err != nil {
		return Cycle{}, Ticket{}, err
	}
	kind, wayfinder := KindTicket, "map"
	created, err := c.CreateTicket(project, TicketInput{Kind: &kind, Title: &title, ParentID: &ticket, Wayfinder: &wayfinder})
	if err != nil {
		return Cycle{}, Ticket{}, err
	}
	cycle, err := c.UpdateCurrentCycle(ticket, CycleInput{MapID: &created.ID})
	if err != nil {
		_ = c.DeleteTicket(created.ID)
		return Cycle{}, Ticket{}, err
	}
	return cycle, created, nil
}
