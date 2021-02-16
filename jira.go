package main

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	jirago "gopkg.in/andygrunwald/go-jira.v1"
)

const jiraDateLayout = "2006/01/02"

type JiraClient struct {
	client     *jirago.Client
	MyUser     *jirago.User
	ProjectKey string
}

func NewJiraClient(baseURL, projectKey, username, token string) *JiraClient {
	tp := jirago.BasicAuthTransport{
		Username: username,
		Password: token,
	}
	client, err := jirago.NewClient(tp.Client(), baseURL)
	if err != nil {
		log.Fatal(err)
	}

	user, res, err := client.User.GetSelf()
	err = jiraResponseErr(res, err)
	if err != nil {
		log.Fatal(fmt.Errorf("error while fetching Jira user (hint: are the username (email) and token provided correct?): %v", err))
	}

	return &JiraClient{
		client:     client,
		ProjectKey: projectKey,
		MyUser:     user,
	}
}

func (c *JiraClient) CreateIssue(date time.Time) (*JiraIssue, error) {
	issue, err := c.createIssue(c.MyUser.AccountID, c.ProjectKey, c.issueName(date))
	if err != nil {
		return nil, err
	}
	if issue == nil {
		log.Fatal("unknown error while creating issue")
	}
	return &JiraIssue{issue}, nil
}

func (c *JiraClient) createIssue(accountID, projectKey, issueName string) (*jirago.Issue, error) {
	issueReq := &jirago.Issue{
		Fields: &jirago.IssueFields{
			Assignee: &jirago.User{
				AccountID: accountID,
			},
			Project: jirago.Project{
				Key: projectKey,
			},
			Summary: issueName,
			Type: jirago.IssueType{
				Name: "Task",
			},
		},
	}
	issue, res, err := c.client.Issue.Create(issueReq)
	err = jiraResponseErr(res, err)
	if err != nil {
		return nil, err
	}
	issue.Fields = issueReq.Fields // quirk: patching fields since they come as nil
	return issue, nil
}

func (c *JiraClient) FindIssue(email string, date time.Time) (issue *JiraIssue, err error) {
	// Query language docs: https://support.atlassian.com/jira-software-cloud/docs/advanced-search-reference-jql-fields/
	projectKey := c.ProjectKey
	issueName := c.issueName(date)
	issueNameHyphenated := strings.ReplaceAll(issueName, "/", "-") // check as well
	jql := fmt.Sprintf(`assignee = "%s" AND (summary ~ "%s" OR summary ~ "%s") AND project = "%s"`, email, issueName, issueNameHyphenated, projectKey)
	issues, res, err := c.client.Issue.Search(jql, nil)
	err = jiraResponseErr(res, err)
	if err != nil {
		return nil, err
	}
	switch len(issues) {
	case 0:
		return nil, fmt.Errorf("found no issue matching asignee='%s' and summary='%s' and project='%s'", email, issueName, projectKey)
	case 1:
		i := issues[0]
		return &JiraIssue{&i}, nil
	default:
		var keys []string
		for _, iss := range issues {
			keys = append(keys, iss.Key)
		}
		return nil, fmt.Errorf("found more than one issue (expected 1) matching email='%s' and date='%s': %s", email, date, strings.Join(keys, ","))
	}
}

// issueName (aka summary) convention for this tool. Used in issue creation and retrieval
func (c *JiraClient) issueName(date time.Time) string {
	return fmt.Sprintf("%s %s", date.Format(jiraDateLayout), date.Weekday().String())
}

type JiraIssue struct {
	*jirago.Issue
}

// LinkToUI of the issue, intended for human consumption (fallbacks to a link to the REST resource)
func (i JiraIssue) LinkToUI() string {
	issueUrl := i.Issue.Self
	u, err := url.Parse(i.Issue.Self)
	if err == nil {
		issueUrl = fmt.Sprintf("%s://%s/browse/%s", u.Scheme, u.Host, i.Issue.Key)
	}
	return issueUrl
}

func jiraResponseErr(res *jirago.Response, err error) error {
	var resDump []byte
	if err != nil {
		var dumpErr error
		if resDump, dumpErr = httputil.DumpResponse(res.Response, true); dumpErr == nil {
			fmt.Printf("------ Jira response: \n%s\n", string(resDump))
		}
		return fmt.Errorf("jira error: %s", err)
	}
	return nil
}
