package main

import (
	"flag"
	"fmt"
	jenk "github.com/NewestUser/jenkins/lib"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"sort"
)

const Version = "0.1"
const CliName = "jenkins"

var commands = map[string]command{
	"jobs":   jobsCmd(),
	"start":  startCmd(),
	"config": configCmd(),
}

func main() {
	fs := flag.NewFlagSet(CliName, flag.ExitOnError)

	version := fs.Bool("version", false, "Print version and exit")

	opts := &jenkinsOpts{
	}

	fs.StringVar(&opts.Host, "h", "", "Host of the jenkins server")
	fs.StringVar(&opts.User, "u", "", "User that the api token belongs to")
	fs.StringVar(&opts.ApiToken, "t", "", "Api token to be used for authentication")

	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), fmt.Sprintf("Usage: %s [global flags] <command> [command flags]", CliName))
		fmt.Fprintf(fs.Output(), "\nglobal flags:\n")
		fs.PrintDefaults()

		names := make([]string, 0, len(commands))
		for name := range commands {
			names = append(names, name)
		}

		sort.Strings(names)
		for _, name := range names {
			if cmd := commands[name]; cmd.fs != nil {
				fmt.Fprintf(fs.Output(), "\n%s command:\n", name)
				cmd.fs.SetOutput(fs.Output())
				cmd.fs.PrintDefaults()
			}
		}
		return
	}

	cmdArgs := os.Args[1:]
	if err := fs.Parse(cmdArgs); err != nil {
		log.Fatal(err)
	}

	if *version {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	jenkinsConf, err := loadJenkinsConfig()
	if err != nil {
		log.Fatalf("failed loading %s, err: %s", JenkinsConfig, err)
	}

	if err := loadOptsFromConf(jenkinsConf, opts); err != nil {
		log.Fatalf("failed reading configuration, err: %s", err)
	}

	for aliasCmdName, aliasCmd := range loadAliasCmds(jenkinsConf) {
		commands[aliasCmdName] = aliasCmd
	}

	if cmd, ok := commands[args[0]]; !ok {
		log.Fatalf("Unknown command: %s", args[0])
	} else if err := cmd.fn(opts, args[1:]); err != nil {
		log.Fatal(err)
	}
}

type command struct {
	fs *flag.FlagSet
	fn func(globalOpts *jenkinsOpts, args []string) error
}

type jenkinsOpts struct {
	Host     string
	User     string
	ApiToken string
}

func newJenkinsClient(globalOpts *jenkinsOpts) (*jenk.Client, error) {
	return jenk.NewJenkinsClient(jenk.Opts{
		Host:     globalOpts.Host,
		User:     globalOpts.User,
		ApiToken: globalOpts.ApiToken,
	})
}

func loadOptsFromConf(iniFile *ini.File, opts *jenkinsOpts) error {
	if len(opts.Host) == 0 {
		opts.Host = loadHostConfig(iniFile)
	}
	if len(opts.ApiToken) == 0 {
		token, err := loadTokenConfig(iniFile)
		if err != nil {
			return err
		}
		opts.ApiToken = token
	}
	if len(opts.User) == 0 {
		opts.User = loadUserConfig(iniFile)
	}

	return nil
}
