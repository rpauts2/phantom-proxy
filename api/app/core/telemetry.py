"""OpenTelemetry - tracing (optional)"""
def setup_telemetry(service_name: str, otlp_endpoint: str | None = None):
    try:
        from opentelemetry import trace
        from opentelemetry.sdk.trace import TracerProvider
        trace.set_tracer_provider(TracerProvider())
        return trace.get_tracer(service_name, "13.0.0")
    except ImportError:
        return None
