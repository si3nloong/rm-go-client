package rm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// GetAccessTokenRequest :
type GetAccessTokenRequest struct {
	GrantType string `json:"grantType"`
}

// GetAccessTokenResponse :
type GetAccessTokenResponse struct {
	AccessToken           string `json:"accessToken"`
	TokenType             string `json:"tokenType"`
	ExpiresIn             int    `json:"expiresIn"`
	RefreshToken          string `json:"refreshToken"`
	RefreshTokenExpiresIn int    `json:"refreshTokenExpiresIn"`
}

func (c *Client) Token() (*oauth2.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == nil {
		if _, err := c.RequestAccessToken(); err != nil {
			return nil, err
		}
	}
	return c.token, nil
}

// RequestAccessToken :
func (c *Client) RequestAccessToken() (*GetAccessTokenResponse, error) {
	src := GetAccessTokenRequest{}
	src.GrantType = "client_credentials"
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	reqUrl, _ := url.Parse(c.oauthEndpoint + "/v1/token")
	req := new(http.Request)
	req.Method = "POST"
	req.URL = reqUrl
	req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(c.clientID+":"+c.clientSecret))},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	dest := GetAccessTokenResponse{}
	if err := json.Unmarshal(b, &dest); err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, &Error{raw: b}
	}

	c.token = &oauth2.Token{
		AccessToken:  dest.AccessToken,
		TokenType:    dest.TokenType,
		RefreshToken: dest.RefreshToken,
		Expiry:       time.Now().UTC().Add(time.Duration(dest.ExpiresIn) * time.Second),
	}
	return &dest, nil
}
