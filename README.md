# Github Action to OTLP

*NOTE: This is still work in progress*

This action outputs Github Action workflows and jobs details to OTLP via gRPC.

## Inputs

## `endpoint`

**Required** The OTLP endpoint which will receive the data.

## `headers`

**Optional** Additional header configuration to pass in as metadata to the gRPC connection.

## Example usage

```
uses: codeboten/github-action-to-otlp@v1
with:
  endpoint: 'grpc.otlpendpoint.io:443'
```
