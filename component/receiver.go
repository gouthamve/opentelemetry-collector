// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component // import "go.opentelemetry.io/collector/component"

import (
	"context"

	"go.opentelemetry.io/collector/consumer"
)

// Deprecated: [v0.67.0] use Config.
type ReceiverConfig = Config

// Deprecated: [v0.67.0] use UnmarshalConfig.
var UnmarshalReceiverConfig = UnmarshalConfig

// A TracesReceiver receives traces.
// Its purpose is to translate data from any format to the collector's internal trace format.
// TracesReceiver feeds a consumer.Traces with data.
//
// For example it could be Zipkin data source which translates Zipkin spans into ptrace.Traces.
type TracesReceiver interface {
	Component
}

// A MetricsReceiver receives metrics.
// Its purpose is to translate data from any format to the collector's internal metrics format.
// MetricsReceiver feeds a consumer.Metrics with data.
//
// For example it could be Prometheus data source which translates Prometheus metrics into pmetric.Metrics.
type MetricsReceiver interface {
	Component
}

// A LogsReceiver receives logs.
// Its purpose is to translate data from any format to the collector's internal logs data format.
// LogsReceiver feeds a consumer.Logs with data.
//
// For example a LogsReceiver can read syslogs and convert them into plog.Logs.
type LogsReceiver interface {
	Component
}

// ReceiverCreateSettings configures Receiver creators.
type ReceiverCreateSettings struct {
	// ID returns the ID of the component that will be created.
	ID ID

	TelemetrySettings

	// BuildInfo can be used by components for informational purposes.
	BuildInfo BuildInfo
}

// ReceiverFactory is factory interface for receivers.
//
// This interface cannot be directly implemented. Implementations must
// use the NewReceiverFactory to implement it.
type ReceiverFactory interface {
	Factory

	// CreateTracesReceiver creates a TracesReceiver based on this config.
	// If the receiver type does not support tracing or if the config is not valid
	// an error will be returned instead.
	CreateTracesReceiver(ctx context.Context, set ReceiverCreateSettings, cfg Config, nextConsumer consumer.Traces) (TracesReceiver, error)

	// TracesReceiverStability gets the stability level of the TracesReceiver.
	TracesReceiverStability() StabilityLevel

	// CreateMetricsReceiver creates a MetricsReceiver based on this config.
	// If the receiver type does not support metrics or if the config is not valid
	// an error will be returned instead.
	CreateMetricsReceiver(ctx context.Context, set ReceiverCreateSettings, cfg Config, nextConsumer consumer.Metrics) (MetricsReceiver, error)

	// MetricsReceiverStability gets the stability level of the MetricsReceiver.
	MetricsReceiverStability() StabilityLevel

	// CreateLogsReceiver creates a LogsReceiver based on this config.
	// If the receiver type does not support the data type or if the config is not valid
	// an error will be returned instead.
	CreateLogsReceiver(ctx context.Context, set ReceiverCreateSettings, cfg Config, nextConsumer consumer.Logs) (LogsReceiver, error)

	// LogsReceiverStability gets the stability level of the LogsReceiver.
	LogsReceiverStability() StabilityLevel
}

// ReceiverFactoryOption apply changes to ReceiverOptions.
type ReceiverFactoryOption interface {
	// applyReceiverFactoryOption applies the option.
	applyReceiverFactoryOption(o *receiverFactory)
}

var _ ReceiverFactoryOption = (*receiverFactoryOptionFunc)(nil)

// receiverFactoryOptionFunc is an ReceiverFactoryOption created through a function.
type receiverFactoryOptionFunc func(*receiverFactory)

func (f receiverFactoryOptionFunc) applyReceiverFactoryOption(o *receiverFactory) {
	f(o)
}

// Deprecated: [v0.67.0] use CreateDefaultConfigFunc.
type ReceiverCreateDefaultConfigFunc = CreateDefaultConfigFunc

// CreateTracesReceiverFunc is the equivalent of ReceiverFactory.CreateTracesReceiver().
type CreateTracesReceiverFunc func(context.Context, ReceiverCreateSettings, Config, consumer.Traces) (TracesReceiver, error)

// CreateTracesReceiver implements ReceiverFactory.CreateTracesReceiver().
func (f CreateTracesReceiverFunc) CreateTracesReceiver(
	ctx context.Context,
	set ReceiverCreateSettings,
	cfg Config,
	nextConsumer consumer.Traces) (TracesReceiver, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateMetricsReceiverFunc is the equivalent of ReceiverFactory.CreateMetricsReceiver().
type CreateMetricsReceiverFunc func(context.Context, ReceiverCreateSettings, Config, consumer.Metrics) (MetricsReceiver, error)

// CreateMetricsReceiver implements ReceiverFactory.CreateMetricsReceiver().
func (f CreateMetricsReceiverFunc) CreateMetricsReceiver(
	ctx context.Context,
	set ReceiverCreateSettings,
	cfg Config,
	nextConsumer consumer.Metrics,
) (MetricsReceiver, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

// CreateLogsReceiverFunc is the equivalent of ReceiverFactory.CreateLogsReceiver().
type CreateLogsReceiverFunc func(context.Context, ReceiverCreateSettings, Config, consumer.Logs) (LogsReceiver, error)

// CreateLogsReceiver implements ReceiverFactory.CreateLogsReceiver().
func (f CreateLogsReceiverFunc) CreateLogsReceiver(
	ctx context.Context,
	set ReceiverCreateSettings,
	cfg Config,
	nextConsumer consumer.Logs,
) (LogsReceiver, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg, nextConsumer)
}

type receiverFactory struct {
	baseFactory
	CreateTracesReceiverFunc
	tracesStabilityLevel StabilityLevel
	CreateMetricsReceiverFunc
	metricsStabilityLevel StabilityLevel
	CreateLogsReceiverFunc
	logsStabilityLevel StabilityLevel
}

func (r receiverFactory) TracesReceiverStability() StabilityLevel {
	return r.tracesStabilityLevel
}

func (r receiverFactory) MetricsReceiverStability() StabilityLevel {
	return r.metricsStabilityLevel
}

func (r receiverFactory) LogsReceiverStability() StabilityLevel {
	return r.logsStabilityLevel
}

// WithTracesReceiver overrides the default "error not supported" implementation for CreateTracesReceiver and the default "undefined" stability level.
func WithTracesReceiver(createTracesReceiver CreateTracesReceiverFunc, sl StabilityLevel) ReceiverFactoryOption {
	return receiverFactoryOptionFunc(func(o *receiverFactory) {
		o.tracesStabilityLevel = sl
		o.CreateTracesReceiverFunc = createTracesReceiver
	})
}

// WithMetricsReceiver overrides the default "error not supported" implementation for CreateMetricsReceiver and the default "undefined" stability level.
func WithMetricsReceiver(createMetricsReceiver CreateMetricsReceiverFunc, sl StabilityLevel) ReceiverFactoryOption {
	return receiverFactoryOptionFunc(func(o *receiverFactory) {
		o.metricsStabilityLevel = sl
		o.CreateMetricsReceiverFunc = createMetricsReceiver
	})
}

// WithLogsReceiver overrides the default "error not supported" implementation for CreateLogsReceiver and the default "undefined" stability level.
func WithLogsReceiver(createLogsReceiver CreateLogsReceiverFunc, sl StabilityLevel) ReceiverFactoryOption {
	return receiverFactoryOptionFunc(func(o *receiverFactory) {
		o.logsStabilityLevel = sl
		o.CreateLogsReceiverFunc = createLogsReceiver
	})
}

// NewReceiverFactory returns a ReceiverFactory.
func NewReceiverFactory(cfgType Type, createDefaultConfig CreateDefaultConfigFunc, options ...ReceiverFactoryOption) ReceiverFactory {
	f := &receiverFactory{
		baseFactory: baseFactory{
			cfgType:                 cfgType,
			CreateDefaultConfigFunc: createDefaultConfig,
		},
	}
	for _, opt := range options {
		opt.applyReceiverFactoryOption(f)
	}
	return f
}
