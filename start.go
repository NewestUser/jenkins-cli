package main

import (
	"flag"
	"fmt"
	"strings"
)

func startCmd() command {
	fs := flag.NewFlagSet("jenkins start", flag.ExitOnError)

	opts := &startOpts{
		jobParams: make(arrayFlag, 0),
	}

	fs.Var(&opts.jobParams, "P", "Job parameter in the format key:value")

	return command{fs: fs, fn: func(globalOpts *jenkinsOpts, args []string) error {
		fs.Parse(args)

		if len(fs.Args()) != 1 {
			return fmt.Errorf("missing job name")
		}

		ctx := &startCtx{jobName: fs.Args()[0]}
		return startJob(ctx, opts, globalOpts)
	}}
}

func startJob(ctx *startCtx, startOpt *startOpts, globalOpts *jenkinsOpts) error {
	jenkins, err := newJenkinsClient(globalOpts)
	if err != nil {
		return err
	}

	jobParams, err := startOpt.params()
	if err != nil {
		return err
	}

	num, err := jenkins.StartJob(ctx.jobName, jobParams)
	if err != nil {
		return err
	}
	reportJobNumber(num)
	return nil
}

type startOpts struct {
	jobParams arrayFlag
}

func (o *startOpts) params() (map[string]string, error) {
	params := make(map[string]string, len(o.jobParams))
	for _, p := range o.jobParams {
		keyAndVal := strings.Split(p, ":")
		if len(keyAndVal) != 2 {
			return nil, fmt.Errorf("invalid parameter %s, format is key:value", p)
		}

		params[keyAndVal[0]] = keyAndVal[1]
	}
	return params, nil
}

type startCtx struct {
	jobName string
}

func reportJobNumber(number int) {
	// TODO (mzlatev) this is not actually the build number
	fmt.Printf("#%d\n", number)
}
