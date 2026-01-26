package config

// OTelConfig contains OpenTelemetry configuration options.
// All standard OpenTelemetry environment variables are also supported via the SDK.
type OTelConfig struct {
	// Enabled enables OpenTelemetry tracing when set to true
	Enabled bool `yaml:"enabled" conf:"default:false"`

	// ServiceName is the name of the service reported to the telemetry backend
	ServiceName string `yaml:"service_name" conf:"default:homebox"`

	// ServiceVersion is the version of the service (defaults to build version)
	ServiceVersion string `yaml:"service_version"`

	// Exporter specifies the exporter type: "otlp", "stdout", or "none"
	Exporter string `yaml:"exporter" conf:"default:otlp"`

	// Endpoint is the OTLP exporter endpoint (e.g., "localhost:4317" for gRPC or "http://localhost:4318" for HTTP)
	// The full address including port should be specified
	Endpoint string `yaml:"endpoint"`

	PathPrefix string `yaml:"path_prefix"`

	// Protocol specifies the OTLP protocol: "grpc" or "http"
	Protocol string `yaml:"protocol" conf:"default:grpc"`

	// Insecure disables TLS for the exporter connection
	Insecure bool `yaml:"insecure" conf:"default:false"`

	// Headers are additional headers to send with OTLP requests (comma-separated key=value pairs)
	Headers string `yaml:"headers"`

	// SampleRate is the sampling rate for traces (0.0 to 1.0, where 1.0 means all traces)
	SampleRate float64 `yaml:"sample_rate" conf:"default:1.0"`

	// EnableMetrics enables OpenTelemetry metrics collection
	EnableMetrics bool `yaml:"enable_metrics" conf:"default:true"`

	// MetricsInterval is the interval at which metrics are exported (e.g., "15s", "1m")
	MetricsInterval string `yaml:"metrics_interval" conf:"default:15s"`

	// EnableLogging enables OpenTelemetry logging
	EnableLogging bool `yaml:"enable_logging" conf:"default:true"`

	// EnableDatabaseTracing enables tracing for database operations
	EnableDatabaseTracing bool `yaml:"enable_database_tracing" conf:"default:true"`

	// EnableHTTPTracing enables tracing for HTTP requests
	EnableHTTPTracing bool `yaml:"enable_http_tracing" conf:"default:true"`

	// ProxyEnabled enables the telemetry proxy endpoint for frontend tracing
	ProxyEnabled bool `yaml:"proxy_enabled" conf:"default:true"`
}
