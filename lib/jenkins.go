package jenk

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type ColorStatus map[string]string

// https://stackoverflow.com/a/43550483/3296947
var jenkinsColors = ColorStatus{
	"red":            "Failed",
	"red_anime":      "In progress",
	"yellow":         "Unstable",
	"yellow_anime":   "In progress",
	"blue":           "Success",
	"blue_anime":     "In progress",
	"grey":           "Pending",
	"grey_anime":     "In progress",
	"disabled":       "Disabled",
	"disabled_anime": "In progress",
	"aborted":        "Aborted",
	"aborted_anime":  "In progress",
	"nobuilt":        "No built",
	"nobuilt_anime":  "In progress",
}

func (b ColorStatus) status(color string) string {
	c := b[strings.ToLower(color)]
	if len(c) == 0 {
		return color
	}
	return c
}

type Opts struct {
	Host     string
	User     string
	ApiToken string
}

func NewJenkinsClient(opts Opts) (*Client, error) {
	jenkins := gojenkins.CreateJenkins(http.DefaultClient, opts.Host, opts.User, opts.ApiToken)
	jenkins, err := jenkins.Init()
	if err != nil {
		return nil, fmt.Errorf("failed initializing jenkins client, err: %s", err)
	}

	return &Client{
		jenkins: jenkins,
	}, nil
}

type Client struct {
	jenkins *gojenkins.Jenkins
}

func (c *Client) GetAllJobs() (*Jobs, error) {
	jobs, err := c.jenkins.GetAllJobNames()
	if err != nil {
		return nil, err
	}

	jobCopy := make([]Job, len(jobs), len(jobs))

	for i, v := range jobs {
		jobCopy[i] = newJob(c.jenkins, v)
	}

	return &Jobs{Entries: jobCopy}, nil
}

func (c *Client) StartJob(jobName string, params map[string]string) (int, error) {
	j, err := c.jenkins.GetJob(jobName)
	if err != nil {
		return 0, err
	}

	jobNumber, err := j.InvokeSimple(params)
	return int(jobNumber), err
}

type Jobs struct {
	Entries []Job
}

func (j *Jobs) FilterByPattern(pattern string, ignoreCase bool) *Jobs {
	if ignoreCase {
		pattern = strings.ToLower(pattern)
	}
	isWildCard := strings.Contains(pattern, "*")

	filteredJob := &Jobs{}
	if isWildCard {
		reg := regexp.MustCompile(strings.ReplaceAll(pattern, "*", ".*"))

		filteredJob.Entries = filter(j.Entries, func(job Job) bool {
			if ignoreCase {
				return reg.MatchString(strings.ToLower(job.Name()))
			}
			return reg.MatchString(job.Name())
		})
	} else {
		filteredJob.Entries = filter(j.Entries, func(job Job) bool {
			if ignoreCase {
				return pattern == strings.ToLower(job.Name())
			}
			return pattern == job.Name()
		})
	}
	return filteredJob
}

func filter(jobs []Job, predicate func(Job) bool) []Job {
	result := make([]Job, 0)
	for _, job := range jobs {
		if predicate(job) {
			result = append(result, job)
		}
	}
	return result
}

type Job interface {
	Name() string
	Status() string
	Url() string
	LastBuildNumber() int64
}

func newJob(client *gojenkins.Jenkins, val gojenkins.InnerJob) *jenkJob {
	return &jenkJob{client: client, value: val}
}

type jenkJob struct {
	client *gojenkins.Jenkins
	value  gojenkins.InnerJob
	job    *gojenkins.Job // lazy init
}

func (j *jenkJob) Name() string {
	return j.value.Name
}

func (j *jenkJob) Status() string {
	return jenkinsColors.status(j.value.Color)
}

func (j *jenkJob) LastBuildNumber() int64 {
	return j.queryJob().Raw.LastBuild.Number
}

func (j *jenkJob) queryJob() *gojenkins.Job {
	if j.job == nil {
		val, err := j.client.GetJob(j.value.Name)
		if err != nil {
			log.Fatalf("failed fetching job %s, err: %s\n", j.value.Name, err)
		}
		j.job = val
	}
	return j.job
}

func (j *jenkJob) Url() string {
	return j.value.Url
}
