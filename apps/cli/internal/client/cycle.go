package client

import "net/http"

type Cycle struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	TicketID   string `json:"ticket_id"`
	Ordinal    int    `json:"ordinal"`
	Phase      string `json:"phase"`
	Resolution string `json:"resolution,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	ClosedAt   string `json:"closed_at,omitempty"`
}

type CycleInput struct {
	Phase      *string `json:"phase,omitempty"`
	Resolution *string `json:"resolution,omitempty"`
}

func (c *Client) ListCycles(ticket string) ([]Cycle, error) {
	var body struct {
		Cycles []Cycle `json:"cycles"`
	}
	if err := c.do(http.MethodGet, "/api/folio/tickets/"+ticket+"/cycles", nil, &body); err != nil {
		return nil, err
	}
	return body.Cycles, nil
}

func (c *Client) OpenCycle(ticket string) (Cycle, error) {
	var cycle Cycle
	err := c.do(http.MethodPost, "/api/folio/tickets/"+ticket+"/cycles", struct{}{}, &cycle)
	return cycle, err
}

func (c *Client) UpdateCycle(id string, in CycleInput) (Cycle, error) {
	var cycle Cycle
	err := c.do(http.MethodPatch, "/api/folio/cycles/"+id, in, &cycle)
	return cycle, err
}
