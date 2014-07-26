package circleci

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	libraryVersion = "0.1"
	defaultBaseURL = "https://circleci.com/api/v1/"
	userAgent      = "go-circleci/" + libraryVersion
)

// A Client manages communication with the CircleCI API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// A CircleCI API token to authenticate requests
	token string

	// Base URL for API requests.  Defaults to the public CircleCI API, but can be
	// set to a domain endpoint to use with CircleCI Enterprise.  BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the CircleCI API.
	UserAgent string
}

// NewClient returns a new CircleCI API client. A token must be provided to
// authenticate API requests. Tokens can be created at https://circleci.com/account/api
func NewClient(token string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{client: http.DefaultClient, token: token, BaseURL: baseURL, UserAgent: userAgent}
}

// NewRequest creates an API request. A relative URL can be provided in path,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	endpoint := c.BaseURL.ResolveReference(rel)

	params, err := url.ParseQuery(endpoint.RawQuery)
	if err != nil {
		return nil, err
	}

	params.Set("circle-token", c.token)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequest(method, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

type Project struct {
	VCSURL   string `json:"vcs_url"`
	Followed bool
	Username string
	RepoName string
	Branches map[string]Branch
}

type Branch struct {
	PusherLogins   []string `json:"pusher_logins"`
	LastNonSuccess Build    `json:"last_non_success"`
	LastSuccess    Build    `json:"last_success"`
	RecentBuilds   []Build  `json:"recent_builds"`
	RunningBuilds  []Build  `json:"running_builds"`
}

type Build struct {
	PushedAt    time.Time `json:"pushed_at"`
	VCSRevision string    `json:"vcs_revision"`
	BuildNum    uint      `json:"build_num"`
	Outcome     string
}

func (c *Client) Projects() ([]Project, error) {
	req, err := c.NewRequest("GET", "projects")
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	} else {
		return projects, nil
	}
}

type DetailedBuild struct {
	VCSURL          string `json:"vcs_url"`
	BuildURL        string `json:"build_url"`
	BuildNum        uint   `json:"build_num"`
	Branch          string
	VCSRevision     string `json:"vcs_revision"`
	CommitterName   string `json:"committer_name"`
	CommitterEmail  string `json:"committer_email"`
	Subject         string
	Body            string
	Why             string
	DontBuild       string    `json:"dont_build"`
	QueuedAt        time.Time `json:"queued_at"`
	StartTime       time.Time `json:"start_time"`
	StopTime        time.Time `json:"stop_time"`
	BuildTimeMillis uint      `json:"build_time_millis"`
	Username        string
	Reponame        string
	Lifecycle       string
	Outcome         string
	Status          string
	RetryOf         uint `json:"retry_of"`
	Previous        struct {
		Status   string
		BuildNum uint `json:"build_num"`
	}
	Steps []Step
}

type Step struct {
	Name    string
	Actions []Action
}

type Action struct {
	BashCommand   string    `json:"bash_command"`
	RunTimeMillis uint      `json:"run_time_millis"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Name          string
	Command       string
	ExitCode      int `json:"exit_code"`
	Type          string
	Index         uint
	Status        string
}