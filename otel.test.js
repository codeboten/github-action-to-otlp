const opentelemetry = require('@opentelemetry/api');
const otel = require('./otel')

test('registerTracer register a global tracer', () => {
    initialTracer = opentelemetry.trace.getTracer('example-basic-tracer-node');
    otel.configureOpenTelemetry('http://localhost:4318/v1/traces');
    newTracer = opentelemetry.trace.getTracer('example-basic-tracer-node');
    expect(newTracer).not.toBe(initialTracer);
});

test('registerTracer validate required input parameters', () => {
    expect(() => otel.configureOpenTelemetry()).toThrow();
});