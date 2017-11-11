package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

func FindInPath(f func(string) (bool, error), path string) (string, error) {
	pi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if pi.IsDir() {
		return _FindInPath(f, path)
	}
	dn, _ := filepath.Split(path)
	return _FindInPath(f, dn)
}

func _FindInPath(f func(string) (bool, error), path string) (string, error) {
	contains, err := f(path)
	if err != nil {
		return "", err
	}
	if contains {
		return path, nil
	}
	dn, fn := filepath.Split(path)
	if fn == "" {
		return "", errors.New("no match found")
	}
	return _FindInPath(f, dn)
}

func IsRcsDir(path string) (bool, error) {
	return DirHasSatisfyingFile(
		func(fi os.FileInfo) bool {
			return fi.Name() == ".git" || fi.Name() == ".svn"
		},
		path)
}

func ContainsVersionFile(path string) (bool, error) {
	return DirHasSatisfyingFile(
		func(fi os.FileInfo) bool {
			return fi.Name() == "version.json"
		},
		path)
}

func DirHasSatisfyingFile(f func(os.FileInfo) bool, path string) (bool, error) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	for _, fi := range fs {
		if f(fi) {
			return true, nil
		}
	}
	return false, nil
}
