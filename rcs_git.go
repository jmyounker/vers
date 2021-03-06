package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

type RcsGit struct {
	Root string
}

func (v RcsGit) Name() string {
	return "git"
}

func (v RcsGit) Branch() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--branch")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return ParseGitStatus(out.String())
}

func (v RcsGit) CommitCounter() (string, error) {
	cmd := exec.Command("git", "rev-list", "HEAD", "--count")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) != 2 {
		return "", errors.New("expected only one line from rev-list")
	}
	// Ensure it can be converted to a number
	c, err := strconv.Atoi(lines[0])
	if err != nil {
		return "", err
	}
	return strconv.Itoa(c), nil
}

func (v RcsGit) RepoCounter() (string, error) {
	return "", errors.New("Git does not support whole-repo commit counters")
}

func (v RcsGit) RepoRoot() (string, error) {
	return "", errors.New("Git does not support repo root")
}

func (v RcsGit) CommitHash() (string, error) {
	cmd := exec.Command("git", "log", "-n", "1", "--pretty=format:%H")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) != 1 {
		return "", errors.New("expected only one line from git log")
	}
	return lines[0], nil
}

func (v RcsGit) CommitHashShort() (string, error) {
	cmd := exec.Command("git", "log", "-n", "1", "--pretty=format:%h")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) != 1 {
		return "", errors.New("expected only one line from git log")
	}
	return lines[0], nil
}

func ParseGitStatus(status string) (string, error) {
	lines := strings.Split(status, "\n")
	if len(lines) == 0 {
		return "", errors.New("expected at least one line of git output")
	}
	branch_line := strings.Split(lines[0], " ")
	if len(branch_line) < 2 {
		return "", errors.New("leading branch line should have at least two elements")
	}
	if branch_line[0] != "##" {
		return "", errors.New("expected line to start with branch marker ##")
	}
	if branch_line[1] == "HEAD" {
		return "HEAD", nil
	}
	tracking_branch_parts := strings.Split(branch_line[1], "...")
	if len(tracking_branch_parts) > 1 {
		return strings.TrimPrefix(tracking_branch_parts[0], "origin/"), nil
	}
	return strings.TrimPrefix(branch_line[1], "origin/"), nil
}
