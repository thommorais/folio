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

type DomainRecord struct {
	ID         string   `json:"id"`
	ClientID   string   `json:"client_id"`
	ClientSlug string   `json:"client_slug"`
	Slug       string   `json:"slug"`
	Name       string   `json:"name"`
	Descr      string   `json:"descr,omitempty"`
	Members    []Member `json:"members"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type DomainInput struct {
	Slug  *string `json:"slug,omitempty"`
	Name  *string `json:"name,omitempty"`
	Descr *string `json:"descr,omitempty"`
}

func (c *Client) ListDomains() ([]DomainRecord, error) {
	var body struct {
		Domains []DomainRecord `json:"domains"`
	}
	if err := c.do(http.MethodGet, "/api/folio/domains", nil, &body); err != nil {
		return nil, err
	}
	return body.Domains, nil
}

func (c *Client) UpdateDomain(clientRef, domainRef string, in DomainInput) (DomainRecord, error) {
	var out DomainRecord
	err := c.do(http.MethodPatch, "/api/folio/clients/"+clientRef+"/domains/"+domainRef, in, &out)
	return out, err
}
