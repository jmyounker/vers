package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	//"github.com/hoisie/mustache"
	//"regexp"
	"regexp"
	"strconv"
)

func actionInit(c *cli.Context) error {
	vf := c.GlobalString("file")
	if (vf == "") {
		return errors.New("version file required")
	}
	return createInitFile(vf)
}

func actionShow(c *cli.Context) error {
	vf := c.GlobalString("file")
	if (vf == "") {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		fmt.Println("Couldnot parse branches")
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := ctx.getBranch()
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

	// get list of unexpanded variables
	vars := format.Variables()
	unexpanded := []string{}
	for _, v := range(vars) {
		_, ok := ctx.State[v]
		if !ok {
			unexpanded = append(unexpanded, v)
		}
	}

	// for each still unexpanded variable, variables perform the expansion
	for _, v := range(unexpanded) {
		value, err := ExpandVariable(v, &ctx)
		if err != nil {
			return err
		}
		ctx.State[v] = value
	}

	// perform expansion
	version, err := format.Expand(&ctx.State)
	if err != nil {
		return nil
	}
	fmt.Println(version)
	return nil
}


func actionDataFile(c *cli.Context) error {
	vf := c.GlobalString("file")
	if (vf == "") {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		fmt.Println("Couldnot parse branches")
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := ctx.getBranch()
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

	// get list of unexpanded variables
	vars := format.Variables()
	unexpanded := []string{}
	for _, v := range(vars) {
		_, ok := ctx.State[v]
		if !ok {
			unexpanded = append(unexpanded, v)
		}
	}

	// for each still unexpanded variable, variables perform the expansion
	for _, v := range(unexpanded) {
		value, err := ExpandVariable(v, &ctx)
		if err != nil {
			return err
		}
		ctx.State[v] = value
	}

	// perform expansion
	version, err := format.Expand(&ctx.State)
	if err != nil {
		return nil
	}
	fmt.Println(version)
	return nil
}



func ExpandVariable(v string, c *Context) (string, error) {
	return "", errors.New(fmt.Sprintf("currently no expansions are defined for %s", v))
}


type Context struct {
	Rcs LazyRcsContainer
	State map[string]string
}

type LazyRcsContainer struct {
	VersionFile string
}


type GenericRcs interface {
	getCommitCounter() int
	getBranch() string
	supportsCommitHashes() bool
	getCommitHash()
}

func actionValidate(c *cli.Context) error {
	vf := c.GlobalString("file")
	if (vf == "") {
		return errors.New("version file required")
	}

	_, err := readConfig(vf)
	if err != nil {
		return err
	}
	return nil
}

func createInitFile(versionFile string) error {
	c := Config{
		Branches: []BranchConfig{{
				BranchPattern: ".*",
				VersionTemplate: "{branch}.{commit-counter}",
			},
		},
	};
	return writeConfig(versionFile, c)
}


func writeConfig(filename string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
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
	for _, bc := range(config.Branches) {
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
		return errors.New(fmt.Sprintf("branch pattern '%s' is malformed", bc.BranchPattern))
	}
	_, err = ParseString(bc.VersionTemplate)
	if err != nil {
		return errors.New(fmt.Sprintf("version template '%s' is malformed", bc.VersionTemplate))
	}
	return nil
}

type Config struct {
	Data DataConfig	`json:"data"`
	Branches []BranchConfig `json:"branches"`
}


type DataConfig struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Release int `json:"release"`
}


type BranchConfig struct {
	BranchPattern string `json:"branch"`
	VersionTemplate string `json:"version"`
}


func (c *Config) getBranchConfig(branch string) (*BranchConfig, error) {
	for _, bc := range(c.Branches) {
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
	Name string
	Value string
}

func NewContext(versionFile string, c *Config, opts []Option) Context {
	ctx := Context{
		Rcs: LazyRcsContainer{
			VersionFile: versionFile,

		},
		State: map[string]string{},
	}
	ctx.State["major"] = strconv.Itoa(c.Data.Major)
	ctx.State["minor"] = strconv.Itoa(c.Data.Minor)
	ctx.State["release"] = strconv.Itoa(c.Data.Release)
	for _, opt := range(opts) {
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
	for _, opt := range(opts) {
		match := optPtrn.FindStringSubmatch(opt)
		if len(match) == 0 {
			return res, errors.New(fmt.Sprintf("cannot parse option '%s'", string(opt)))
		}
		o := Option{
			Name: string(match[1]),
			Value: string(match[2]),
		}
		res = append(res, o)
	}
	return res, nil
}


func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Usage:       "Version file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "init",
			Action: actionInit,
		},
		{
			Name: "test-config",
			Action: actionValidate,
		},
		{
			Name: "show",
			Action: actionShow,
			Flags: []cli.Flag {
				cli.StringSliceFlag{
					Name: "option, X",
					Usage: "Specified option",
				},
			},
		},
		{
			Name: "data-file",
			Action: actionDataFile,
			Flags: []cli.Flag {
				cli.StringSliceFlag{
					Name: "option, X",
					Usage: "Specified option",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}




//func prepareConfig(config Json, state map[string]string) map[string]string;
//func prepareOpts(opt Options, state map[string]string) map[string]string;
//func prepareRcs(rcs Rcs, state map[string]string) map[string]string;

//func createRcs(ConfigFile string) (error, *Rcs);

//func createDataFile(config json, state map[string]string);
//func createVersion(format Template, state map[string]string) (error, string);
