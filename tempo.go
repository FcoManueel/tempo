package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const tempoAPI = "https://api.tempo.io/core/3"
const tempoDateLayout = "2006-01-02"

// TempoClient for tempo API defined in https://apidocs.tempo.io/
type TempoClient struct {
	token         string
	jiraAccountID string
}

// NewTempoClient returns a tempo API Client
func NewTempoClient(token, jiraAccountID string) *TempoClient {
	return &TempoClient{
		token:         token,
		jiraAccountID: jiraAccountID,
	}
}

func (c *TempoClient) Do(method, resource string, body io.Reader) (*http.Response, error) {
	wrap := func(err error) error {
		return errors.Wrap(err, fmt.Sprintf("while doing request (%s %s)", method, resource))
	}

	if !strings.HasPrefix(resource, "/") {
		resource = "/" + resource
	}

	endpoint := fmt.Sprintf("%s%s", tempoAPI, resource)
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	if debug {
		reqDump, _ := httputil.DumpRequest(req, true)
		fmt.Printf("------ Tempo request: \n%s\n", string(reqDump))
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, wrap(err)
	}

	success := res.StatusCode >= 200 && res.StatusCode < 300
	if !success {
		resDump, _ := httputil.DumpResponse(res, true)
		if debug {
			fmt.Printf("------ Tempo response: \n%s\n", string(resDump))
		}
		err = fmt.Errorf("unsuccessful status (%d). \nResponse: \n%s", res.StatusCode, string(resDump))
		return nil, wrap(err)
	}
	return res, nil
}

func (c *TempoClient) LogDay(date time.Time, hours int, jiraIssueKey string) error {
	loggedSeconds := hours * 60 * 60

	w := NewWorklog{
		IssueKey:                 jiraIssueKey,
		TimeSpentSeconds:         loggedSeconds,
		BillableSeconds:          loggedSeconds,
		StartDate:                date.Format(tempoDateLayout),
		StartTime:                "00:00:00",
		Description:              fmt.Sprintf("Working on issue %s.", jiraIssueKey),
		AuthorAccountID:          c.jiraAccountID,
		RemainingEstimateSeconds: 0,
		Attributes:               nil,
	}

	body, err := json.Marshal(w)
	if err != nil {
		return err
	}
	_, err = c.Do("POST", "/worklogs", bytes.NewReader(body))
	return err
}

func (c *TempoClient) GetLoggedHours(jiraIssueKey string) (int, error) {
	endpoint := fmt.Sprintf("/worklogs?issue=%s", jiraIssueKey)
	res, err := c.Do("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}

	var worklogs WorklogsRes
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&worklogs); err != nil {
		return 0, err
	}
	loggedSeconds := 0
	for _, worklog := range worklogs.Results {
		loggedSeconds += worklog.BillableSeconds
	}
	return loggedSeconds / 60 / 60, nil
}

type NewWorklog struct {
	IssueKey                 string `json:"issueKey"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds"`
	BillableSeconds          int    `json:"billableSeconds"`
	StartDate                string `json:"startDate"`
	StartTime                string `json:"startTime"`
	Description              string `json:"description"`
	AuthorAccountID          string `json:"authorAccountId"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds"`
	Attributes               []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes,omitempty"`
}

type Worklog struct {
	Self           string `json:"self"`
	TempoWorklogID int    `json:"tempoWorklogId"`
	JiraWorklogID  int    `json:"jiraWorklogId"`
	Issue          struct {
		Self string `json:"self"`
		Key  string `json:"key"`
		ID   int    `json:"id"`
	} `json:"issue"`
	TimeSpentSeconds int       `json:"timeSpentSeconds"`
	BillableSeconds  int       `json:"billableSeconds"`
	StartDate        string    `json:"startDate"`
	StartTime        string    `json:"startTime"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Author           struct {
		Self        string `json:"self"`
		AccountID   string `json:"accountId"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
	Attributes struct {
		Self   string        `json:"self"`
		Values []interface{} `json:"values"`
	} `json:"attributes"`
}

type WorklogsRes struct {
	Self     string `json:"self"`
	Metadata struct {
		Count  int `json:"count"`
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"metadata"`
	Results []Worklog `json:"results"`
}
