package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"path/filepath"
	"regexp"
	"strconv"
)

var version string;

func actionInit(c *cli.Context) error {
	vf := c.GlobalString("file")
	if vf == "" {
		return errors.New("version file required")
	}
	tn := c.String("template")
	if tn == "" {
		tn = "default"
	}
	return createInitFile(vf, tn, c.String("rcs"))
}

var InitTemplates = map[string]Config{
	"default": Config{
		Data: map[string]interface{}{},
		Branches: []BranchConfig{{
			BranchPattern:   ".*",
			VersionTemplate: "{branch}.{commit-counter}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"version",
		},
	},
	"semvar": Config{
		Data: map[string]interface{}{
			"major":   0,
			"minor":   0,
			"release": 1,
		},
		Branches: []BranchConfig{{
			BranchPattern:   ".*",
			VersionTemplate: "{major}.{minor}.{release}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"major",
			"minor",
			"release",
			"version",
		},
	},
	"python": Config{
		Data: map[string]interface{}{
			"major":   0,
			"minor":   0,
			"release": 1,
		},
		Branches: []BranchConfig{{
			BranchPattern:   "master|trunk",
			VersionTemplate: "{major}.{minor}.{release}",
		},{
			BranchPattern:   ".*",
			VersionTemplate: "{major}.{minor}.{release}dev{commit-counter}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"major",
			"minor",
			"release",
			"version",
		},
	},
}

func actionShow(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		return errors.New(fmt.Sprintf("Could not parse branches"))
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := LookupParameter("branch", &ctx)
	if err != nil {
		return err
	}

	// locate appropriate branch config
	// if branch does not match, error
	branchConfig, err := config.getBranchConfig(branch)
	if err != nil {
		return err
	}

	format, err := ParseString(branchConfig.VersionTemplate)
	if err != nil {
		return err
	}

	// perform expansion
	version, err := format.Expand(&ctx)
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}

func actionDataFile(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	df := c.String("data-file")

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		return errors.New(fmt.Sprintf("Could not parse branches"))
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := LookupParameter("branch", &ctx)
	if err != nil {
		return err
	}

	// locate appropriate branch config
	// if branch does not match, error
	branchConfig, err := config.getBranchConfig(branch)
	if err != nil {
		return err
	}

	format, err := ParseString(branchConfig.VersionTemplate)
	if err != nil {
		return err
	}

	// perform expansion
	version, err := format.Expand(&ctx)
	if err != nil {
		return err
	}
	ctx.State["version"] = version

	data := map[string]string{}
	for _, v := range ctx.Config.DataFileFields {
		value, err := LookupParameter(v, &ctx)
		if err != nil {
			return err
		}
		data[v] = value
	}

	if df == "" {
		d, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", d)
		return nil
	}
	err = writeDataFile(df, data)
	if err != nil {
		return err
	}
	return nil
}

func actionBumpMajor(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	major, err := config.GetDataInt("major")
	if err != nil {
		return err
	}

	// Ensure that minor is defined
	_, err = config.GetDataInt("minor")
	if err != nil {
		return err
	}

	// Ensure that release is defined
	_, err = config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["major"] = major + 1
	config.Data["minor"] = 0
	config.Data["release"] = 0

	err = writeConfig(vf, *config)
	if err != nil {
		return err
	}
	return nil
}

func actionBumpMinor(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	minor, err := config.GetDataInt("minor")
	if err != nil {
		return err
	}

	// Ensure that release is defined in the file
	_, err = config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["minor"] = minor + 1
	config.Data["release"] = 0

	err = writeConfig(vf, *config)
	if err != nil {
		return err
	}
	return nil
}

func actionBumpRelease(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	release, err := config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["release"] = release + 1

	err = writeConfig(vf, *config)
	if err != nil {
		return err
	}
	return nil
}

type Context struct {
	VersionFile string
	Rcs         Rcs
	State       map[string]string
	Config      Config
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

func actionValidate(c *cli.Context) error {
	vf := c.GlobalString("file")
	if vf == "" {
		return errors.New("version file required")
	}

	_, err := readConfig(vf)
	if err != nil {
		return err
	}
	return nil
}

func createInitFile(versionFile string, templateName string, rcsName string) error {
	c, ok := InitTemplates[templateName]
	if !ok {
		return fmt.Errorf("unnknown template: %s", templateName)
	}
	if rcsName == "" {
		rcs, err := GetRcs(filepath.Dir(versionFile))
		if err == nil {
			rcsName = rcs.Name()
		}
	}
	for _, df := range(RcsDataFileFields(rcsName)) {
		c.DataFileFields = append(c.DataFileFields, df)
	}
	return writeConfig(versionFile, c)
}

func RcsDataFileFields(rcsName string) []string {
	switch rcsName {
	case "git":
		return []string{
			"commit-hash",
			"commit-hash-short",
		}
	case "svn":
		return []string{
			"repo-counter",
			"repo-root",
		}
	default:
		return []string{}
	}
}

func writeConfig(filename string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0664)
}

func writeDataFile(filename string, dataFile map[string]string) error {
	data, err := json.MarshalIndent(dataFile, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0664)
}

func readConfig(filename string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if len(config.Branches) == 0 {
		return nil, errors.New("confing must contain at least one branch expressions")
	}
	for _, bc := range config.Branches {
		err := checkBranchConfig(bc)
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func checkBranchConfig(bc BranchConfig) error {
	if bc.BranchPattern == "" {
		return errors.New("branch pattern required")
	}
	if bc.VersionTemplate == "" {
		return errors.New("version template required")
	}
	_, err := regexp.Compile(bc.BranchPattern)
	if err != nil {
		return fmt.Errorf("branch pattern '%s' is malformed", bc.BranchPattern)
	}
	t, err := ParseString(bc.VersionTemplate)
	if err != nil {
		return fmt.Errorf("version template '%s' is malformed", bc.VersionTemplate)
	}
	err = ValidateTemplateAsVersion(t)
	if err != nil {
		return err
	}
	return nil
}

type Config struct {
	Data           map[string]interface{} `json:"data"`
	Branches       []BranchConfig         `json:"branches"`
	DataFileFields []string               `json:"data-file"`
}

//type DataConfig struct {
//	Major   int `json:"major"`
//	Minor   int `json:"minor"`
//	Release int `json:"release"`
//}
//
type BranchConfig struct {
	BranchPattern   string `json:"branch"`
	VersionTemplate string `json:"version"`
}

func (c *Config) getBranchConfig(branch string) (*BranchConfig, error) {
	for _, bc := range c.Branches {
		ptrn := regexp.MustCompile(bc.BranchPattern)
		if ptrn.MatchString(branch) {
			return &bc, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no branch config matching branch '%s'", branch))
}

func (c *Context) getBranch() (string, error) {
	branch, ok := c.State["branch"]
	if ok {
		return branch, nil
	}
	return "", errors.New("branch from RCS not yet implemented")
}

type Option struct {
	Name  string
	Value string
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

func getOptions(c *cli.Context) ([]Option, error) {
	opts := c.StringSlice("option")
	res := []Option{}
	if opts == nil {
		return res, nil
	}
	optPtrn := regexp.MustCompile("([^=]+)=(.*)")
	for _, opt := range opts {
		match := optPtrn.FindStringSubmatch(opt)
		if len(match) == 0 {
			return res, errors.New(fmt.Sprintf("cannot parse option '%s'", string(opt)))
		}
		o := Option{
			Name:  string(match[1]),
			Value: string(match[2]),
		}
		res = append(res, o)
	}
	return res, nil
}

func ValidateTemplateAsVersion(t Template) error {
	for _, v := range t.Variables() {
		if v == "version" {
			return errors.New("{version} cannot be contained in the version template")
		}
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Version file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "init",
			Action: actionInit,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "template",
					Usage: "Choose a version file template",
				},
				cli.StringFlag{
					Name:  "rcs",
					Usage: "Override RCS specification",
				},
			},
		},
		{
			Name:   "test-config",
			Action: actionValidate,
		},
		{
			Name:   "show",
			Action: actionShow,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "option, X",
					Usage: "Specified option",
				},
			},
		},
		{
			Name:   "data-file",
			Action: actionDataFile,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "option, X",
					Usage: "Specified option",
				},
				cli.StringFlag{
					Name:  "data-file, o",
					Usage: "Data file",
				},
			},
		},
		{
			Name:   "bump-major",
			Action: actionBumpMajor,
		},
		{
			Name:   "bump-minor",
			Action: actionBumpMinor,
		},
		{
			Name:   "bump-release",
			Action: actionBumpRelease,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetVersionFile(c *cli.Context) (string, error) {
	rf := c.GlobalString("file")
	if rf == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", errors.New(fmt.Sprintf("could not locate version file: %s", err.Error()))
		}
		dn, err := FindInPath(ContainsVersionFile, wd)
		if err != nil {
			return "", errors.New(fmt.Sprintf("could not locate version file: %s", err.Error()))
		}
		return filepath.Join(dn, "version.json"), nil
	}
	vf, err := filepath.Abs(filepath.Clean(rf))
	if err != nil {
		return "", err
	}
	return vf, nil
}

func LookupParameter(parameter string, c *Context) (string, error) {
	// Pull from state first, getting memoized or hard-coded values.
	v, ok := c.State[parameter]
	if ok {
		return v, nil
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

func (c *Config) HasData(name string) bool {
	_, ok := c.Data[name]
	return ok
}

func (c *Config) GetDataInt(name string) (int, error) {
	v, ok := c.Data[name]
	if !ok {
		return 0, fmt.Errorf("data field '%s' is not defined", name)
	}
	switch v.(type) {
	case int:
		return v.(int), nil
	case float64:
		return int(v.(float64)), nil
	case string:
		iv, err := strconv.Atoi(v.(string))
		if err != nil {
			return 0, fmt.Errorf("cannot convert '%s' to an int: %s", name, err.Error())
		}
		return iv, nil
	default:
		return 0, fmt.Errorf("'%s' is not an int", name)
	}
}

func (c *Config) GetDataString(name string) (string, error) {
	v, ok := c.Data[name]
	if !ok {
		return "", fmt.Errorf("data field '%s' is not defined", name)
	}
	switch v.(type) {
	case int:
		return strconv.Itoa(v.(int)), nil
	case float64:
		return strconv.Itoa(int(v.(float64))), nil
	case string:
		return v.(string), nil
	default:
		return "", fmt.Errorf("expected '%s' to be a string", name)
	}
}

var ParameterLookups = map[string]func(c *Context) (string, error){
	"branch":            LookupBranch,
	"commit-counter":    LookupCommitCounter,
	"repo-counter":      LookupRepoCounter,
	"commit-hash":       LookupCommitHash,
	"commit-hash-short": LookupCommitHashShort,
	"repo-root":         LookupRepoRoot,
}

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
