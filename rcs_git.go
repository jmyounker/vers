package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type RcsGit struct {
	Root string
}

func (v RcsGit) Branch() (string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--branch")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) == 0 {
		return "", errors.New("expected at least one line of git output")
	}
	branch_line := strings.Split(lines[0], " ")
	if len(branch_line) != 2 {
		return "", errors.New("leading branch line should have at least two elements")
	}
	if branch_line[0] != "##" {
		return "", errors.New("expected line to start with branch marker ##")
	}
	return branch_line[1], nil
}

func (v RcsGit) CommitCounter() (string, error) {
	return "", nil
}

func (v RcsGit) CommitHash() (string, error) {
	return "", nil
}
