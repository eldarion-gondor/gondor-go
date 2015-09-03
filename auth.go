package gondor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) EnsureAuthed() (bool, error) {
	var err error
	c.checkAuthOnce.Do(func() {
		if c.opts.Auth.AccessToken == "" {
			err = fmt.Errorf("auth: access token not present")
			return
		}
		var req *http.Request
		var resp *http.Response
		url := c.buildBaseURL("/")
		req, err = http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.opts.Auth.AccessToken))
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return
		}
		switch resp.StatusCode {
		case 200:
			c.authed = true
			return
		case 401:
			err = c.AuthenticateWithRefreshToken()
			if err == nil {
				c.authed = true
			}
			break
		default:
			err = fmt.Errorf("auth: unhandled status code returned (%s)", resp.Status)
		}
	})
	return c.authed, err
}

func (c *Client) Authenticate(username, password string) (*ClientOpts, error) {
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/token/",
		url.Values{
			"grant_type": {"password"},
			"client_id":  {c.opts.ID},
			"username":   {username},
			"password":   {password},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		return nil, errors.New("authentication failed")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != "" {
		return nil, fmt.Errorf("authentication request failed: %q", payload.ErrorDescription)
	}
	c.opts.Auth.Username = username
	c.opts.Auth.AccessToken = payload.AccessToken
	c.opts.Auth.RefreshToken = payload.RefreshToken
	return c.opts, nil
}

func (c *Client) AuthenticateWithRefreshToken() error {
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/token/",
		url.Values{
			"grant_type":    {"refresh_token"},
			"client_id":     {c.opts.ID},
			"refresh_token": {c.opts.Auth.RefreshToken},
		},
	)
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		return errors.New("authentication failed")
	}
	var payload struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Error != "" {
		return fmt.Errorf("authentication request failed: %q", payload.ErrorDescription)
	}
	c.opts.Auth.AccessToken = payload.AccessToken
	c.opts.Auth.RefreshToken = payload.RefreshToken
	return nil
}

func (c *Client) RevokeAccess() (*ClientOpts, error) {
	resp, err := http.PostForm(
		"https://identity.gondor.io/oauth/revoke_token/",
		url.Values{
			"client_id": {c.opts.ID},
			"token":     {c.opts.Auth.RefreshToken},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to log out (%s)", resp.Status)
	}
	c.opts.Auth.Username = ""
	c.opts.Auth.AccessToken = ""
	c.opts.Auth.RefreshToken = ""
	return c.opts, nil
}
