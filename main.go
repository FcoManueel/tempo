package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	version      = "1.0"
	defaultHours = 8
	debug        = false
)

func main() {
	app := cli.NewApp()
	app.Name = "tempo"
	app.Version = version
	app.Usage = "log worked hours using Jira and the Tempo plugin"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:   "log",
			Usage:  "create a Jira ticket and assign it worked hours in Tempo",
			Action: ActionLog,
		},
		{
			Name:   "see",
			Usage:  "check if there exists a Jira ticket and a Tempo entry",
			Action: ActionSee,
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "jira-url",
			Usage:    "base url for your company Jira (e.g. https://my-company.atlassian.net/ )",
			Required: true,
			EnvVars:  []string{"JIRA_URL"},
		},
		&cli.StringFlag{
			Name:     "jira-project-key",
			Usage:    "key for Jira project used to track timesheet tasks",
			Required: true,
			EnvVars:  []string{"JIRA_PROJECT_KEY"},
		},
		&cli.StringFlag{
			Name:     "jira-user",
			Usage:    "jira username (email)",
			Required: true,
			EnvVars:  []string{"JIRA_USERNAME"},
		},
		&cli.StringFlag{
			Name:     "jira-token",
			Usage:    "jira REST API token",
			Required: true,
			EnvVars:  []string{"JIRA_TOKEN"},
		},
		&cli.StringFlag{
			Name:     "tempo-token",
			Usage:    "tempo REST API token",
			Required: true,
			EnvVars:  []string{"TEMPO_TOKEN"},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func ActionLog(c *cli.Context) error {
	timesheet := NewTimesheet(c)
	dates, hours := parseArgs(c)
	for _, date := range dates {
		err := timesheet.Log(date, hours)
		if err != nil {
			// Don't fail the whole thing, just log the error
			fmt.Printf("Failed attempt to log %d hours for %s %s: %v\n", hours, date.Weekday(), date, err)
		}
	}

	fmt.Println("Done. Have a nice day!")
	return nil
}

func ActionSee(c *cli.Context) error {
	timesheet := NewTimesheet(c)
	dates, _ := parseArgs(c)
	for _, date := range dates {
		err := timesheet.See(date)
		if err != nil {
			// Don't fail the whole thing, just log the error
			fmt.Printf("Failed attempt to see logged hours for %s %s: %v\n", date.Weekday(), date, err)
		}
	}
	return nil
}

func parseArgs(c *cli.Context) (dates []time.Time, hours int) {
	const (
		day   = 24 * time.Hour
		week  = "week"
		today = "today"
	)

	dateArg := c.Args().First()
	switch {
	case strings.HasPrefix(dateArg, week): // e.g dateArg="week-2"
		// Handle +/- modifiers
		modifier := strings.TrimPrefix(dateArg, week) // e.g modifier="-2"
		offset, _ := strconv.Atoi(modifier)

		// set date to this week's monday
		date := time.Now()
		for date.Weekday() != time.Monday {
			date = date.Add(-1 * day)
		}
		// apply offset
		date = date.Add(time.Duration(offset) * 7 * day)

		// add 5 days starting from monday
		for i := 0; i < 5; i++ {
			t := date.Add(time.Duration(i) * day)
			dates = append(dates, t)
		}
	case strings.HasPrefix(dateArg, today): // e.g dateArg="today+1"
		// Handle +/- modifiers
		modifier := strings.TrimPrefix(dateArg, today) // e.g modifier="+1"
		offset, _ := strconv.Atoi(modifier)

		// set date to today and apply offset
		date := time.Now().Add(time.Duration(offset) * day)
		dates = append(dates, date)
	default:
		date, err := time.Parse("2006/01/02", dateArg)
		if err != nil {
			date, err = time.Parse("2006-01-02", dateArg) // support hyphenated input as well
			if err != nil {
				log.Fatal("unrecognized argument for date: ", dateArg)
			}
		}
		dates = append(dates, date)
	}

	var err error
	hours = defaultHours
	tail := c.Args().Tail()
	if len(tail) > 1 {
		hours, err = strconv.Atoi(tail[0])
		if err != nil {
			log.Fatal("unrecognized argument for hours", tail[0])
		}
	}
	return dates, hours
}
