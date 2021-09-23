# Github Action to OTLP

This action outputs Github Action workflows and jobs details to OTLP.

## Inputs

## `endpoint`

**Required** The OTLP endpoint which will receive the data.

## Outputs

## `time`

The time we greeted you.

## Example usage

uses: actions/github-action-to-otlp@v1
with:
  endpoint: 'grpc.otlpendpoint.io:443'
