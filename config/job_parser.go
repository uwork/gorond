package config

import (
	"errors"
	"fmt"
	"gorond/util"
	"regexp"
	"strings"
)

// Job構造体
type Job struct {
	Line     string
	Schedule string
	User     string
	Command  string
	Indent   int
	Status   string
	Childs   []*Job
	Parent   *Job
}

const (
	WAITING = "waiting"
	RUNNING = "running"
	FAILED  = "failed"
)

// cronコマンドパーサ正規表現
var lineValidator1 = regexp.MustCompile(`^\t+`)
var lineParser1 = regexp.MustCompile(`^(@[^ ]+? +)((?:[0-9]+[dhms])+ +)?([a-zA-Z_\-$]+) +(.+)$`)
var lineParser2 = regexp.MustCompile(`^((?:(?:[0-9\*\/\,\-\?]+|[A-Z]+) +){6})([a-zA-Z_\-$]+) +(.+)$`)
var lineParser3 = regexp.MustCompile(`^( +)- +([a-zA-Z_\-$]+) +(.+)$`)
var commentRegex = regexp.MustCompile(`^((?:[^#'"]|"[^"]*"|'[^']*')*)(#.*)?$`)

// Job設定をパースする。
func parseJobConfig(str string) ([]*Job, error) {
	jobs := make([]*Job, 0, 10)

	var last *Job = nil
	lines := strings.Split(str, "\n")
	for idx, line := range lines {
		// コメントを除外
		if commentRegex.MatchString(line) {
			line = commentRegex.ReplaceAllString(line, "$1")
		}

		if util.Trim(line) == "" {
			continue
		}

		job, err := parseJobLine(line)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)

		// インデント幅に応じて、ツリー構造を構築する。
		if last != nil && job.Indent > 0 {
			if last.Indent == job.Indent {
				job.Parent = last.Parent
				job.Parent.Childs = append(job.Parent.Childs, &job)
			} else if last.Indent < job.Indent {
				job.Parent = last
				job.Parent.Childs = append(job.Parent.Childs, &job)
			} else {
				return nil, errors.New(fmt.Sprintf("job invalid indent level( %d: %s)\n", idx, line))
			}
		}
		last = &job
	}

	return jobs, nil
}

// Job設定1行をパースする。
func parseJobLine(line string) (Job, error) {
	job := Job{Status: WAITING}
	job.Line = line

	var result [][]string

	// タブが含まれていないかチェック
	if lineValidator1.MatchString(line) {
		return Job{}, errors.New(fmt.Sprintf("'%s' include tab.", line))
	}

	// フォーマットに合わせて読み出す
	if line[0:1] == "@" {
		result = lineParser1.FindAllStringSubmatch(line, -1)
		if 0 == len(result) {
			return Job{}, errors.New(fmt.Sprintf("parse error (%s).", line))
		}
		schedule := util.Trim(util.Trim(result[0][1]) + " " + util.Trim(result[0][2]))
		user := util.Trim(result[0][3])
		command := util.Trim(strings.Join(result[0][4:], " "))

		job.Schedule = schedule
		job.User = user
		job.Command = command
	} else if util.Trim(line)[0:1] == "-" {
		result = lineParser3.FindAllStringSubmatch(line, -1)
		if 0 == len(result) {
			return Job{}, errors.New(fmt.Sprintf("parse error (%s).", line))
		}
		indentSize := len(result[0][1])
		user := util.Trim(result[0][2])
		command := util.Trim(strings.Join(result[0][3:], " "))

		job.Indent = indentSize
		job.User = user
		job.Command = command
	} else {
		result = lineParser2.FindAllStringSubmatch(line, -1)
		if 0 == len(result) {
			return Job{}, errors.New(fmt.Sprintf("parse error (%s).", line))
		}
		schedule := util.Trim(result[0][1])
		user := util.Trim(result[0][2])
		command := util.Trim(strings.Join(result[0][3:], " "))

		job.Schedule = schedule
		job.User = user
		job.Command = command
	}

	return job, nil
}
