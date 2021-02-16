package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

type Timesheet struct {
	tempo *TempoClient
	jira  *JiraClient
}

func NewTimesheet(ctx *cli.Context) Timesheet {
	jiraUrl := ctx.String("jira-url")
	jiraUser := ctx.String("jira-user")
	jiraToken := ctx.String("jira-token")
	tempoToken := ctx.String("tempo-token")
	projectKey := ctx.String("jira-project-key")

	if !strings.HasSuffix(jiraUrl, "/") {
		jiraUrl += "/"
	}

	jira := NewJiraClient(jiraUrl, projectKey, jiraUser, jiraToken)
	tempo := NewTempoClient(tempoToken, jira.MyUser.AccountID)

	return Timesheet{
		tempo: tempo,
		jira:  jira,
	}
}

func (t *Timesheet) Log(date time.Time, hours int) error {
	issue, err := t.jira.CreateIssue(date)
	if err != nil {
		return err
	}
	t.printDetail(date, issue, hours)
	if err := t.tempo.LogDay(date, hours, issue.Key); err != nil {
		return err
	}
	return nil
}

func (t *Timesheet) See(date time.Time) error {
	issue, err := t.jira.FindIssue(t.jira.MyUser.EmailAddress, date)
	if err != nil {
		return err
	}
	hours, err := t.tempo.GetLoggedHours(issue.Key)
	if err != nil {
		return err
	}
	t.printDetail(date, issue, hours)
	return nil
}

func (t *Timesheet) printDetail(date time.Time, issue *JiraIssue, hours int) {
	if issue == nil {
		fmt.Printf("no Jira issue found for %s\n", date)
	} else {
		fmt.Printf("%s  %dh  %s\n", issue.LinkToUI(), hours, issue.Fields.Summary)
	}
}
