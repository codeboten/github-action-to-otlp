package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type actionConfig struct {
	workflow         string
	githubRepository string
	owner            string
	repo             string
	runID            string
}

// TODO: add attributes using https://docs.github.com/en/actions/learn-github-actions/environment-variables
// TODO: add user-agent
// TODO: add support for auth

func getSteps(ctx context.Context, conf actionConfig) error {
	tracer := otel.Tracer(conf.githubRepository)
	client := github.NewClient(nil)
	id, err := strconv.ParseInt(conf.runID, 10, 64)
	if err != nil {
		return err
	}
	workflow, _, err := client.Actions.GetWorkflowRunByID(ctx, conf.owner, conf.repo, id)
	if err != nil {
		return err
	}

	ctx, workflowSpan := tracer.Start(ctx, *workflow.Name, trace.WithTimestamp(workflow.CreatedAt.Time))
	defer workflowSpan.End(trace.WithTimestamp(workflow.UpdatedAt.Time))

	jobs, _, err := client.Actions.ListWorkflowJobs(ctx, conf.owner, conf.repo, id, &github.ListWorkflowJobsOptions{})
	if err != nil {
		return err
	}

	for _, job := range jobs.Jobs {
		ctx, jobSpan := tracer.Start(ctx, *job.Name, trace.WithTimestamp(job.GetStartedAt().Time))
		if err != nil {
			return err
		}
		for _, step := range job.Steps {
			_, stepSpan := tracer.Start(ctx, *step.Name, trace.WithTimestamp(step.GetStartedAt().Time))
			if step.CompletedAt != nil {
				stepSpan.End(trace.WithTimestamp(step.CompletedAt.Time))
			} else {
				stepSpan.End()
			}
		}
		if job.CompletedAt != nil {
			jobSpan.End(trace.WithTimestamp(job.CompletedAt.Time))
		} else {
			jobSpan.End()
		}

	}

	return nil
}

func parseConfig() (actionConfig, error) {
	githubRepository, ok := os.LookupEnv("GITHUB_REPOSITORY")
	if !ok {
		return actionConfig{}, errors.New("missing variable: GITHUB_REPOSITORY")
	}

	runID, ok := os.LookupEnv("GITHUB_RUN_ID")
	if !ok {
		return actionConfig{}, errors.New("missing variable: GITHUB_RUN_ID")
	}

	workflowName, ok := os.LookupEnv("GITHUB_WORKFLOW")
	if !ok {
		return actionConfig{}, errors.New("missing variable: GITHUB_WORKFLOW")
	}

	parts := strings.Split(githubRepository, "/")
	if len(parts) < 2 {
		return actionConfig{}, fmt.Errorf("invalid variable GITHUB_REPOSITORY: %s", githubRepository)
	}
	conf := actionConfig{
		workflow:         workflowName,
		githubRepository: githubRepository,
		owner:            parts[0],
		repo:             parts[1],
		runID:            runID,
	}

	return conf, nil
}

func main() {

	conf, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}
	lsOtel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName(conf.githubRepository),
	)
	defer lsOtel.Shutdown()

	if err != nil {
		log.Printf("%v", err)
	}

	err = getSteps(context.Background(), conf)
	if err != nil {
		log.Println(err)
	}
}
