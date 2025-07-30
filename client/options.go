package client

import "net/http"

type Option func(*Client)

func WithHTTPClient(cl *http.Client) Option {
	return func(c *Client) {
		c.httpClient = cl
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithAuthURL(authURL string) Option {
	return func(c *Client) {
		c.authURL = authURL
	}
}
