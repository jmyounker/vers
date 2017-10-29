package main

import (
	"errors"
	"io/ioutil"
)

func GetRcs(versionFile string) (Rcs, error) {
	dn, err := FindInPath(IsRcsDir, versionFile)
	if err != nil {
		return nil, err
	}
	fis, err := ioutil.ReadDir(dn)
	if err != nil {
		return nil, err
	}
	for _, fi := range fis {
		if fi.Name() == ".git" {
			return RcsGit{Root: dn}, nil
		}
		if fi.Name() == ".svn" {
			return RcsSvn{Root: dn}, nil
		}
	}
	return nil, errors.New("could not locates RCS root containing version file")
}

type Rcs interface {
	Branch() (string, error)
	CommitCounter() (string, error)
	RepoCounter() (string, error)
	RepoRoot() (string, error)
	CommitHash() (string, error)
	CommitHashShort() (string, error)
}
