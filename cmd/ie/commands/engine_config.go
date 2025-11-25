package commands

import "github.com/Azure/InnovationEngine/internal/engine"

var buildEngineConfiguration = func(opts *executionOptions, overrides ...func(*engine.EngineConfiguration)) engine.EngineConfiguration {
	cfg := engine.EngineConfiguration{
		Verbose:          opts.Verbose,
		DoNotDelete:      opts.DoNotDelete,
		StreamOutput:     opts.StreamOutput,
		Subscription:     opts.Subscription,
		CorrelationId:    opts.CorrelationID,
		Environment:      opts.Environment,
		WorkingDirectory: opts.WorkingDirectory,
		RenderValues:     opts.RenderValues,
		ReportFile:       opts.ReportFile,
	}

	for _, override := range overrides {
		if override != nil {
			override(&cfg)
		}
	}

	return cfg
}
