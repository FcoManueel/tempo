# Tempo logger
This is a tool to help with the repetitive task of opening Jira issues and
using the Tempo plugin to log worked hours every day.

It was designed under the assumption that you want to manage your logs keeping a
**1-1 relationship between worked days and Jira tickets**.


## Installation

```bash
go get github.com/FcoManueel/tempo 
```

## Usage

The `tempo` tool has two commands, `tempo see` and `tempo log`, explained below. All date formats are `YYYY/MM/DD`. For usage help just type `tempo`.

*Notes:   
The following examples implicitly use the environment variables `$JIRA_URL`, `$JIRA_PROJECT_KEY`,`$JIRA_USERNAME`,`$JIRA_TOKEN`, and`$TEMPO_TOKEN`.  
For information on how to provide that data through flags run `tempo help` and look for "global options".  
For information on how to generate the required tokens go to the last section of this readme.*

### Log

To log 8-hours of work for a given date:
```bash
tempo log 2021/01/31
```

You can send `log` a second parameter indicating the amount of hours to log :
```bash
tempo log 2021/01/31 4
```

You can use `today` instead of today's date, and `week` to refer to this week's Mon—Fri.
```bash
tempo log today
tempo log week 
```

Lastly, you can use `+#`, `-#` suffixes for relative dates.
```bash
tempo log today-1   // yesterday
tempo log today+1   // tomorrow
tempo log week-3    // three week ago
```

Example: *Log 4 hours of work for each workday of last week*
```bash
tempo log week-1 4
``` 

### See
If you want to inspect a particular date for entries created by this tool:
```bash
tempo see 2021/02/08
```

Syntactic sugar supported by `log` is supported by this command as well:
```bash
tempo see week-1
> https://foo.atlassian.net/browse/BAR-365  8h  2021/02/08 Monday
> https://foo.atlassian.net/browse/BAR-366  8h  2021/02/09 Tuesday
> https://foo.atlassian.net/browse/BAR-367  8h  2021/02/10 Wednesday
> https://foo.atlassian.net/browse/BAR-368  8h  2021/02/11 Thursday
> https://foo.atlassian.net/browse/BAR-369  8h  2021/02/12 Friday
```

# Useful info
### How to create a Jira Token  

- Go to https://id.atlassian.com/manage/api-tokens using your Jira account.
- Click **Create API token** and assign it a name.
- Store the token locally before closing the modal.

### How to create a Tempo Token  

- Go to the Tempo settings on your Jira domain (e.g. https://`<company>`.atlassian.net/plugins/servlet/ac/io.tempo.jira/tempo-app#!/configuration/api-integration ). If you have trouble finding them check the [Tempo docs](https://apidocs.tempo.io/) for more info).
- Click **+ New Token** and assign it a name and durability.
  - Choose your token privileges. Minimum required is `Worklogs` scope.
- Store the token locally before closing the modal.

[comment]: <> (### Potential improvements)
[comment]: <> (- A `delete` command to remove logged hours)
[comment]: <> (- Idempotency for `log` command)