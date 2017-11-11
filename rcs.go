package main

import (
	"errors"
	"io/ioutil"
	"os"
)

func GetRcs(versionFile string) (Rcs, error) {
	_, ok := os.LookupEnv("TRAVIS_BRANCH")
	if ok {
		return RcsTravis{}, nil
	}
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
	Name() string
	Branch() (string, error)
	CommitCounter() (string, error)
	RepoCounter() (string, error)
	RepoRoot() (string, error)
	CommitHash() (string, error)
	CommitHashShort() (string, error)
}
