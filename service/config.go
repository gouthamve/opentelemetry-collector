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

package service // import "go.opentelemetry.io/collector/service"

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/telemetry"
)

var (
	errMissingExporters        = errors.New("no enabled exporters specified in config")
	errMissingReceivers        = errors.New("no enabled receivers specified in config")
	errMissingServicePipelines = errors.New("service must have at least one pipeline")
)

// Config defines the configuration for the various elements of collector or agent.
type Config struct {
	// Receivers is a map of ComponentID to Receivers.
	Receivers map[component.ID]component.Config

	// Exporters is a map of ComponentID to Exporters.
	Exporters map[component.ID]component.Config

	// Processors is a map of ComponentID to Processors.
	Processors map[component.ID]component.Config

	// Extensions is a map of ComponentID to extensions.
	Extensions map[component.ID]component.Config

	Service ConfigService
}

// Validate returns an error if the config is invalid.
//
// This function performs basic validation of configuration. There may be more subtle
// invalid cases that we currently don't check for but which we may want to add in
// the future (e.g. disallowing receiving and exporting on the same endpoint).
func (cfg *Config) Validate() error {
	// Currently, there is no default receiver enabled.
	// The configuration must specify at least one receiver to be valid.
	if len(cfg.Receivers) == 0 {
		return errMissingReceivers
	}

	// Validate the receiver configuration.
	for recvID, recvCfg := range cfg.Receivers {
		if err := component.ValidateConfig(recvCfg); err != nil {
			return fmt.Errorf("receiver %q has invalid configuration: %w", recvID, err)
		}
	}

	// Currently, there is no default exporter enabled.
	// The configuration must specify at least one exporter to be valid.
	if len(cfg.Exporters) == 0 {
		return errMissingExporters
	}

	// Validate the exporter configuration.
	for expID, expCfg := range cfg.Exporters {
		if err := component.ValidateConfig(expCfg); err != nil {
			return fmt.Errorf("exporter %q has invalid configuration: %w", expID, err)
		}
	}

	// Validate the processor configuration.
	for procID, procCfg := range cfg.Processors {
		if err := component.ValidateConfig(procCfg); err != nil {
			return fmt.Errorf("processor %q has invalid configuration: %w", procID, err)
		}
	}

	// Validate the extension configuration.
	for extID, extCfg := range cfg.Extensions {
		if err := component.ValidateConfig(extCfg); err != nil {
			return fmt.Errorf("extension %q has invalid configuration: %w", extID, err)
		}
	}

	return cfg.validateService()
}

func (cfg *Config) validateService() error {
	// Check that all enabled extensions in the service are configured.
	for _, ref := range cfg.Service.Extensions {
		// Check that the name referenced in the Service extensions exists in the top-level extensions.
		if cfg.Extensions[ref] == nil {
			return fmt.Errorf("service references extension %q which does not exist", ref)
		}
	}

	// Must have at least one pipeline.
	if len(cfg.Service.Pipelines) == 0 {
		return errMissingServicePipelines
	}

	// Check that all pipelines have at least one receiver and one exporter, and they reference
	// only configured components.
	for pipelineID, pipeline := range cfg.Service.Pipelines {
		if pipelineID.Type() != component.DataTypeTraces && pipelineID.Type() != component.DataTypeMetrics && pipelineID.Type() != component.DataTypeLogs {
			return fmt.Errorf("unknown pipeline datatype %q for %v", pipelineID.Type(), pipelineID)
		}

		// Validate pipeline has at least one receiver.
		if len(pipeline.Receivers) == 0 {
			return fmt.Errorf("pipeline %q must have at least one receiver", pipelineID)
		}

		// Validate pipeline receiver name references.
		for _, ref := range pipeline.Receivers {
			// Check that the name referenced in the pipeline's receivers exists in the top-level receivers.
			if cfg.Receivers[ref] == nil {
				return fmt.Errorf("pipeline %q references receiver %q which does not exist", pipelineID, ref)
			}
		}

		// Validate pipeline processor name references.
		procSet := make(map[component.ID]bool, len(cfg.Processors))
		for _, ref := range pipeline.Processors {
			// Check that the name referenced in the pipeline's processors exists in the top-level processors.
			if cfg.Processors[ref] == nil {
				return fmt.Errorf("pipeline %q references processor %q which does not exist", pipelineID, ref)
			}
			// Ensure no processors are duplicated within the pipeline
			if _, exists := procSet[ref]; exists {

				return fmt.Errorf("pipeline %q references processor %q multiple times", pipelineID, ref)
			}
			procSet[ref] = true
		}

		// Validate pipeline has at least one exporter.
		if len(pipeline.Exporters) == 0 {
			return fmt.Errorf("pipeline %q must have at least one exporter", pipelineID)
		}

		// Validate pipeline exporter name references.
		for _, ref := range pipeline.Exporters {
			// Check that the name referenced in the pipeline's Exporters exists in the top-level Exporters.
			if cfg.Exporters[ref] == nil {
				return fmt.Errorf("pipeline %q references exporter %q which does not exist", pipelineID, ref)
			}
		}

		if err := cfg.Service.Telemetry.Validate(); err != nil {
			fmt.Printf("telemetry config validation failed, %v\n", err)
		}
	}
	return nil
}

// ConfigService defines the configurable components of the service.
type ConfigService struct {
	// Telemetry is the configuration for collector's own telemetry.
	Telemetry telemetry.Config `mapstructure:"telemetry"`

	// Extensions are the ordered list of extensions configured for the service.
	Extensions []component.ID `mapstructure:"extensions"`

	// Pipelines are the set of data pipelines configured for the service.
	Pipelines map[component.ID]*ConfigServicePipeline `mapstructure:"pipelines"`
}

type ConfigServicePipeline = config.Pipeline
