package main

import (
	"flag"
	"fmt"
	jenk "github.com/NewestUser/jenkins/lib"
)

func jobsCmd() command {
	fs := flag.NewFlagSet(fmt.Sprintf("%s jobs", CliName), flag.ExitOnError)

	opts := &jobsOpts{
	}

	fs.BoolVar(&opts.detailed, "d", false, "Show job details")
	fs.BoolVar(&opts.ignoreCase, "i", true, "Ignore case")

	return command{fs, func(globalOpts *jenkinsOpts, args []string) error {
		fs.Parse(args)

		ctx := &jobCtx{}
		if len(fs.Args()) > 0 {
			ctx.jobName = fs.Args()[0]
		}

		return listAvailableJobs(globalOpts, opts, ctx)
	}}
}

type jobsOpts struct {
	detailed   bool
	ignoreCase bool
}

type jobCtx struct {
	jobName string
}

func (ctx jobCtx) isEmpty() bool {
	return len(ctx.jobName) == 0
}

func listAvailableJobs(globalOpts *jenkinsOpts, jobsOpts *jobsOpts, ctx *jobCtx) error {
	jenkins, err := newJenkinsClient(globalOpts)
	if err != nil {
		return err
	}

	jobs, err := jenkins.GetAllJobs()
	if err != nil {
		return fmt.Errorf("failed to retrieve job names, err: %s", err)
	}

	if !ctx.isEmpty() {
		jobs = jobs.FilterByPattern(ctx.jobName, jobsOpts.ignoreCase)
	}

	reportJobs(jobs, jobsOpts.detailed)
	return nil
}

func reportJobs(jobs *jenk.Jobs, detailed bool) {
	var pad = 0
	for _, job := range jobs.Entries {
		if pad < len(job.Name()) {
			pad = len(job.Name())
		}
	}

	formatter := PrettyFormat{}

	formatter.PadField(jobs.Entries, func(i int) interface{} {
		return jobs.Entries[i].Name()
	})
	formatter.Append("\t")
	formatter.PadField(jobs.Entries, func(i int) interface{} {
		return jobs.Entries[i].Status()
	})

	if detailed {
		formatter.Append("\t#").PadField(jobs.Entries, func(i int) interface{} {
			return jobs.Entries[i].LastBuildNumber()
		}).Append("\t%s")
	}

	format := formatter.Append("\n").Format()
	for _, job := range jobs.Entries {
		if detailed {
			fmt.Printf(format, job.Name(), job.Status(),  job.LastBuildNumber(), job.Url())
		} else {
			fmt.Printf(format, job.Name(), job.Status())
		}
	}
}
