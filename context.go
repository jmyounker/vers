package main

import (
	"fmt"
	"os"
	"strings"
)

type Context struct {
	VersionFile  string
	Rcs          Rcs
	State        map[string]string
	Config       Config
	BranchParams map[string]string
	BranchConfig *BranchConfig
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

func LookupParameter(parameter string, c *Context) (string, error) {
	// Provides memoization/calculate once semantics for param lookup.
	// Memoization is important because some derived values may change
	// between requests.  We want to ensure that only a single value
	// is ever used. Others, such as RCS operations, can potentially
	// be expensive, so we wan't to avoid re-running them if possible.
	v, ok := c.State[parameter]
	if ok {
		return v, nil
	}
	v, err := LookupParameterWithoutMemoization(parameter, c)
	if err != nil {
		return v, err
	}
	c.State[parameter] = v
	return v, nil
}

func LookupParameterWithoutMemoization(parameter string, c *Context) (string, error) {
	// Parameter lookup is a layer cake of sources.  The general
	// idea is that user input should override other values.
	//   * Command line flags override everything.
	//   * Environment variables with exact name match (build-id)
	//   * Environment variables with convention match (BUILD_ID)
	//   * Values derived from the branch name.
	//   * Calculated values from VCS or other sources.
	//   * Values from the config data section.
	//  Having the config data section last lets it function as s
	//  source of default values.
	//
	// Check the environment for overrides.  First we check the raw
	// parameter name.
	ev, ok := os.LookupEnv(parameter)
	if ok {
		return ev, nil
	}
	// If it's missing then we look for a envar-ish looking name
	// variant.  E.g. instead of commit-counter we look for
	// COMMIT_COUNTER.
	ev, ok = os.LookupEnv(MakeEnvarName(parameter))
	if ok {
		return ev, nil
	}
	// Next we see if the parameter could be found in the branch name.
	// Minus signs are illegal in the regex matching group names, so
	// we translate them to underscores.
	bp, ok := c.BranchParams[strings.Replace(parameter, "-", "_", -1)]
	if ok {
		return bp, nil
	}
	// Next we see if it can be calculated.
	f, ok := ParameterLookups[parameter]
	if ok {
		return f(c)
	}
	// Now get defaults from the data sections.
	v, ok := c.BranchConfig.Data[parameter]
	if ok {
		return ParamDataToString(v)
	}
	// Finally we look for values supplied in the config's data section.
	v, ok = c.Config.Data[parameter]
	if ok {
		return ParamDataToString(v)
	}
	return "", fmt.Errorf("unknown parameter %s", parameter)
}

var ParameterLookups = map[string]func(c *Context) (string, error){
	"branch":            LookupBranch,
	"commit-counter":    LookupCommitCounter,
	"repo-counter":      LookupRepoCounter,
	"commit-hash":       LookupCommitHash,
	"commit-hash-short": LookupCommitHashShort,
	"repo-root":         LookupRepoRoot,
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

func MakeEnvarName(s string) string {
	upper := strings.ToUpper(s)
	return strings.Replace(upper, "-", "_", -1)
}
