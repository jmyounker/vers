package main

import (
	"os/exec"
	"bytes"
	"strings"
	"errors"
	"regexp"
	"strconv"
	"encoding/xml"
)

type RcsSvn struct {
	Root string
}

func (v RcsSvn) Name() string {
	return "svn"
}

func (v RcsSvn) Branch() (string, error) {
	info, err := v.SvnInfo()
	if err != nil {
		return "", nil
	}
	url, ok := info["Repository Root"]
	if !ok {
		return "", errors.New("could not find URL in svn output")
	}
	return ParseBranchFromSvnUrl(url)
}

func (v RcsSvn) CommitCounter() (string, error) {
	cmd := exec.Command("svn", "log", "-l", "1", "--xml")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return ParseRevisionFromXmlLog(out.String())
}

func (v RcsSvn) RepoCounter() (string, error) {
	info, err := v.SvnInfo()
	if err != nil {
		return "", nil
	}
	rev, ok := info["Revision"]
	if !ok {
		return "", errors.New("could not find revision in svn output")
	}
	n, err := strconv.Atoi(rev)
	if err != nil {
		return "", errors.New("could not read revision as number")
	}
	return strconv.Itoa(n), nil
}

func (v RcsSvn) RepoRoot() (string, error) {
	info, err := v.SvnInfo()
	if err != nil {
		return "", nil
	}
	url, ok := info["Repository Root"]
	if !ok {
		return "", errors.New("could not find repo root in svn output")
	}
	return url, nil
}

func (v RcsSvn) CommitHash() (string, error) {
	return "", errors.New("SVN does not support commit hashes")
}

func (v RcsSvn) CommitHashShort() (string, error) {
	return "", errors.New("SVN does not support commit hashes")
}

func (v RcsSvn) SvnInfo() (map[string]string, error) {
	cmd := exec.Command("svn", "info")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return ParseSvnInfo(out.String())
}

func ParseSvnInfo(svnOut string) (map[string]string, error) {
	lines := strings.Split(svnOut, "\n")
	if len(lines) == 0 {
		return nil, errors.New("expected at least one line of svn info output")
	}
	info := map[string]string{}
	for _, line := range(lines) {
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			return nil, errors.New("found unparsable svn info line")
		}
		info[kv[0]] = kv[1]
	}
	return info, nil
}

func ParseBranchFromSvnUrl(url string) (string, error) {
	ptrns := []*regexp.Regexp{
		regexp.MustCompile("/branches/([^/]+)"),
		regexp.MustCompile("/(trunk)/"),
		regexp.MustCompile("/(trunk)$"),
		regexp.MustCompile("/tags/([^/]+)"),
	}
	for _, ptrn := range(ptrns) {
		m := ptrn.FindStringSubmatch(url)
		if len(m) == 2 {
			return string(m[1]), nil
		}
	}
	return "", errors.New("could not extract branch from svn URL")
}

type LogRecord struct {
	XMLName xml.Name `xml:"log"`
	LogEntry LogEntry `xml:"logentry"`
}

type LogEntry struct {
	Revision int `xml:"revision,attr"`
}

func ParseRevisionFromXmlLog(log string) (string, error) {
	lr := LogRecord{}
	err := xml.Unmarshal([]byte(log), &lr)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(lr.LogEntry.Revision), nil
}
