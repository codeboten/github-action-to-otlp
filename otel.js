'use strict';

const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { BasicTracerProvider, ConsoleSpanExporter, SimpleSpanProcessor } = require('@opentelemetry/sdk-trace-base');
const { OTLPTraceExporter } =  require('@opentelemetry/exporter-trace-otlp-proto');

function configureOpenTelemetry(endpoint, headers, serviceName) {
    if (!endpoint) {
        throw new Error('endpoint is required');
    }
    if (!serviceName) {
      serviceName = 'github-action-to-otlp';
    }
    const provider = new BasicTracerProvider({
      resource: new Resource({
        [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
      }),
    });
      
    const exporter = new OTLPTraceExporter({
      url: endpoint,
      headers: headers,
      concurrencyLimit: 1,
    });
    provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
    provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
    provider.register();
    return exporter
}

exports.configureOpenTelemetry = configureOpenTelemetry;