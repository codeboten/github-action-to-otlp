package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/lightstep/otel-launcher-go/pipelines"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
)

type actionConfig struct {
	workflow         string
	githubRepository string
	owner            string
	repo             string
	runID            string
	pipelineConfig   pipelines.PipelineConfig
}

// TODO: add attributes using https://docs.github.com/en/actions/learn-github-actions/environment-variables
// TODO: add user-agent
// TODO: add support for auth

func getSteps(ctx context.Context, conf actionConfig) error {
	tracer := otel.Tracer(conf.githubRepository)
	client := github.NewClient(nil)

	// login using the GITHUB_TOKEN coming from the jobs
	// as per https://docs.github.com/en/actions/security-guides/automatic-token-authentication
	githubToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

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

// Code inspired from the opentelemetry-go OTLP exporter
//
// https://github.com/open-telemetry/opentelemetry-go/blob/92551d3933c9c7ef5eaf4f93f876a5487d0024b9/exporters/otlp/otlpmetric/internal/otlpconfig/envconfig.go#L172
func stringToHeader(value string) map[string]string {
	headersPairs := strings.Split(value, ",")
	headers := make(map[string]string)

	for _, header := range headersPairs {
		nameValue := strings.SplitN(header, "=", 2)
		if len(nameValue) < 2 {
			continue
		}
		name, err := url.QueryUnescape(nameValue[0])
		if err != nil {
			continue
		}
		trimmedName := strings.TrimSpace(name)
		trimmedValue := strings.TrimSpace(nameValue[1])

		headers[trimmedName] = trimmedValue
	}

	return headers
}

func parseConfig() (actionConfig, error) {
	endpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok || len(endpoint) == 0 {
		return actionConfig{}, errors.New("invalid endpoint")
	}

	headers := stringToHeader(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))

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

	attributes := []attribute.KeyValue{
		attribute.String(conventions.AttributeServiceName, githubRepository),
	}

	r, _ := resource.New(context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(attributes...),
	)

	insecure := false
	conf := actionConfig{
		workflow:         workflowName,
		githubRepository: githubRepository,
		owner:            parts[0],
		repo:             parts[1],
		runID:            runID,
		pipelineConfig: pipelines.PipelineConfig{
			Endpoint:    endpoint,
			Insecure:    insecure, // TODO: provide config for this
			Headers:     headers,
			Propagators: []string{"tracecontext"}, // TODO: provide config for this
			Resource:    r,
		},
	}

	return conf, nil
}

func main() {

	conf, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	pipelineShutdown, err := pipelines.NewTracePipeline(conf.pipelineConfig)
	defer pipelineShutdown()

	if err != nil {
		log.Printf("%v", err)
	}

	err = getSteps(context.Background(), conf)
	if err != nil {
		log.Println(err)
	}
}
