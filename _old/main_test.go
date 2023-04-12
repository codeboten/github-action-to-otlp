package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {

	_, err := parseConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid endpoint")

	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "endpoint:443")

	_, err = parseConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing variable: GITHUB_REPOSITORY")

	os.Setenv("GITHUB_REPOSITORY", "garbage")

	_, err = parseConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing variable: GITHUB_RUN_ID")

	os.Setenv("GITHUB_RUN_ID", "123")

	_, err = parseConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing variable: GITHUB_WORKFLOW")

	os.Setenv("GITHUB_WORKFLOW", "test name")

	_, err = parseConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid variable GITHUB_REPOSITORY: garbage")

	os.Setenv("GITHUB_REPOSITORY", "test/code")
	conf, err := parseConfig()
	require.NoError(t, err)
	require.Equal(t, "test", conf.owner)
	require.Equal(t, "code", conf.repo)
}
