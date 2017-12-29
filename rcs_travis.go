package main

import (
	"errors"
	"fmt"
	"os"
)

type RcsTravis struct {
	Rcs Rcs
}

func (v RcsTravis) Name() string {
	return "travis"
}

func (v RcsTravis) Branch() (string, error) {
	pb, ok := os.LookupEnv("TRAVIS_PULL_REQUEST_BRANCH")
	if ok && pb != "" {
		return pb, nil
	}
	b, ok := os.LookupEnv("TRAVIS_BRANCH")
	if ok {
		return b, nil
	}
	return "", errors.New("cannot locate branch in environment")
}

func (v RcsTravis) CommitCounter() (string, error) {
	return "UNKNOWN", nil
}

func (v RcsTravis) RepoCounter() (string, error) {
	return "", errors.New("Travis-git does not support whole-repo commit counters")
}

func (v RcsTravis) RepoRoot() (string, error) {
	return "", errors.New("Travis-git does not support repo root")
}

func (v RcsTravis) CommitHash() (string, error) {
	pn, ok := os.LookupEnv("TRAVIS_PULL_REQUEST_NUMBER")
	if ok && pn != "false" {
		return pn, nil
	}
	c, ok := os.LookupEnv("TRAVIS_COMMIT")
	if ok {
		return c, nil
	}
	return "", errors.New("cannot find commit hash in environment")
}

func (v RcsTravis) CommitHashShort() (string, error) {
	c, err := v.CommitHash()
	if err != nil {
		return "", err
	}
	if len(c) < 7 {
		return "", fmt.Errorf("malformed commit hash: '%s'", c)
	}
	return c[0:7], nil
}
