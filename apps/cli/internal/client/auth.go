package client

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Session struct {
	Token string
	Email string
	Name  string
}

// Bypasses do: this is the call that establishes the token.
func (c *Client) Login(identity, password string) (Session, error) {
	payload, err := json.Marshal(map[string]string{"identity": identity, "password": password})
	if err != nil {
		return Session{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/collections/users/auth-with-password", bytes.NewReader(payload))
	if err != nil {
		return Session{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return Session{}, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 {
		failure := apiError{status: res.StatusCode}
		_ = json.NewDecoder(res.Body).Decode(&failure)
		return Session{}, failure
	}

	var body struct {
		Token  string `json:"token"`
		Record struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"record"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return Session{}, err
	}

	return Session{Token: body.Token, Email: body.Record.Email, Name: body.Record.Name}, nil
}
