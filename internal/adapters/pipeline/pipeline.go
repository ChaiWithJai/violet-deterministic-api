package pipeline

import "context"

// Stage is a deterministic pipeline stage. Stages must be pure with respect
// to input payload + explicit external dependencies.
type Stage func(ctx context.Context, payload map[string]any) (map[string]any, error)

func Run(ctx context.Context, initial map[string]any, stages ...Stage) (map[string]any, error) {
	payload := initial
	var err error
	for _, stage := range stages {
		payload, err = stage(ctx, payload)
		if err != nil {
			return nil, err
		}
	}
	return payload, nil
}
