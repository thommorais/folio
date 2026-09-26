package client

import "net/http"

type ClientRecord struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Site      string `json:"site,omitempty"`
	Logo      string `json:"logo,omitempty"`
	Descr     string `json:"descr,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ClientInput struct {
	Slug  *string `json:"slug,omitempty"`
	Name  *string `json:"name,omitempty"`
	Site  *string `json:"site,omitempty"`
	Logo  *string `json:"logo,omitempty"`
	Descr *string `json:"descr,omitempty"`
}

func (c *Client) ListClients() ([]ClientRecord, error) {
	var body struct {
		Clients []ClientRecord `json:"clients"`
	}
	if err := c.do(http.MethodGet, "/api/folio/clients", nil, &body); err != nil {
		return nil, err
	}
	return body.Clients, nil
}

func (c *Client) UpdateClient(ref string, in ClientInput) (ClientRecord, error) {
	var out ClientRecord
	err := c.do(http.MethodPatch, "/api/folio/clients/"+ref, in, &out)
	return out, err
}
