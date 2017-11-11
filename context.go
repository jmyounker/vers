package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Context struct {
	VersionFile string
	Rcs         Rcs
	State       map[string]string
	Config      Config
}

func NewContext(versionFile string, c *Config, opts []Option) Context {
	ctx := Context{
		VersionFile: versionFile,
		State:       map[string]string{},
		Config:      *c,
	}
	for _, opt := range opts {
		ctx.State[opt.Name] = opt.Value
	}
	return ctx
}

func (c *Context) GetRcs() (Rcs, error) {
	if c.Rcs != nil {
		return c.Rcs, nil
	}
	rcs, err := GetRcs(c.VersionFile)
	if err != nil {
		return nil, err
	}
	c.Rcs = rcs
	return rcs, nil
}

func (c *Context) getBranch() (string, error) {
	branch, ok := c.State["branch"]
	if ok {
		return branch, nil
	}
	return "", errors.New("branch from RCS not yet implemented")
}

func LookupParameter(parameter string, c *Context) (string, error) {
	// Pull from state first, getting memoized or hard-coded values.
	v, ok := c.State[parameter]
	if ok {
		return v, nil
	}
	// Next we check the environment for overrides.  First we
	// check the raw parameter name.
	ev, ok := os.LookupEnv(parameter)
	if ok {
		c.State[parameter] = ev
		return ev, nil
	}
	// If it's missing then we look for a envar-ish looking name
	// variant.  E.g. instead of commit-counter we look for
	// COMMIT_COUNTER.
	ev, ok = os.LookupEnv(MakeEnvarName(parameter))
	if ok {
		c.State[parameter] = ev
		return ev, nil
	}
	// Next we look for values supplied in the config's data section.
	if c.Config.HasData(parameter) {
		v, err := c.Config.GetDataString(parameter)
		if err != nil {
			return "", err
		}
		return v, nil
	}
	// Finally we check to see if this is something that can be calculated.
	f, ok := ParameterLookups[parameter]
	if !ok {
		return "", errors.New(fmt.Sprintf("unknown parameter %s", parameter))
	}
	v, err := f(c)
	if err != nil {
		return "", err
	}
	// We have a value, so we memoize it, ensuring that subsequent calls
	// get the same value *and* don't have to calculate it.
	c.State[parameter] = v
	return v, nil
}

func LookupFromRcs(c *Context, f func(Rcs) (string, error)) (string, error) {
	rcs, err := c.GetRcs()
	if err != nil {
		return "", err
	}
	cc, err := f(rcs)
	if err != nil {
		return "", err
	}
	return cc, nil
}

func LookupBranch(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.Branch() })
}

func LookupCommitCounter(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.CommitCounter() })
}

func LookupRepoCounter(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.RepoCounter() })
}

func LookupCommitHash(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.CommitHash() })
}

func LookupCommitHashShort(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.CommitHashShort() })
}

func LookupRepoRoot(c *Context) (string, error) {
	return LookupFromRcs(c, func(r Rcs) (string, error) { return r.RepoRoot() })
}

var ParameterLookups = map[string]func(c *Context) (string, error){
	"branch":            LookupBranch,
	"commit-counter":    LookupCommitCounter,
	"repo-counter":      LookupRepoCounter,
	"commit-hash":       LookupCommitHash,
	"commit-hash-short": LookupCommitHashShort,
	"repo-root":         LookupRepoRoot,
}

func MakeEnvarName(s string) string {
	upper := strings.ToUpper(s)
	return strings.Replace(upper, "-", "_", -1)
}