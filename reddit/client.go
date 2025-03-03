package reddit

import (
	"bytes"
	"fmt"
	"net/http"
)

// tokenURL is the url of reddit's oauth2 authorization service.
const tokenURL = "https://www.reddit.com/api/v1/access_token"

// clientConfig holds all the information needed to define Client behavior, such
// as who the client will identify as externally and where to authorize.
type clientConfig struct {
	// Agent is the user agent set in all requests made by the Client.
	agent string

	// If all fields in App are set, this client will attempt to identify as
	// a registered Reddit app using the credentials.
	app App

	// Custom http client, if nil default should be used
	client *http.Client
}

// client executes http Requests and invisibly handles OAuth2 authorization.
type client interface {
	Do(*http.Request) ([]byte, error)
}

type baseClient struct {
	cli *http.Client
}

func (b *baseClient) Do(req *http.Request) ([]byte, error) {
	resp, err := b.cli.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden:
		return nil, ErrPermissionDenied
	case http.StatusServiceUnavailable:
		return nil, ErrBusy
	case http.StatusTooManyRequests:
		return nil, ErrRateLimit
	case http.StatusBadGateway:
		return nil, ErrGateway
	case http.StatusGatewayTimeout:
		return nil, ErrGatewayTimeout
	default:
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// newClient returns a new client using the given user to make requests.
func newClient(c clientConfig) (client, error) {
	if c.app.tokenURL == "" {
		c.app.tokenURL = tokenURL
	}

	if c.app.unauthenticated() {
		return &baseClient{clientWithAgent(c.agent)}, nil
	}

	if err := c.app.validateAuth(); err != nil {
		return nil, err
	}

	return newAppClient(c)
}
