package providers

import "context"

// ConfigRecord represents a single provider/tenant configuration entry.
type ConfigRecord struct {
	Provider string
	Tenant   string
	Config   Config
}

// ConfigLoader exposes provider configuration data to bootstrap registries.
type ConfigLoader interface {
	Load(ctx context.Context) ([]ConfigRecord, error)
}

// StaticConfigLoader provides a simple in-memory implementation of ConfigLoader.
type StaticConfigLoader struct {
	records []ConfigRecord
}

// NewStaticConfigLoader constructs a StaticConfigLoader from a nested map.
func NewStaticConfigLoader(cfg map[string]map[string]Config) *StaticConfigLoader {
	records := make([]ConfigRecord, 0)
	for provider, tenants := range cfg {
		for tenant, config := range tenants {
			records = append(records, ConfigRecord{
				Provider: provider,
				Tenant:   tenant,
				Config:   config,
			})
		}
	}
	return &StaticConfigLoader{records: records}
}

// Load returns all configuration records stored in the loader.
func (l *StaticConfigLoader) Load(ctx context.Context) ([]ConfigRecord, error) {
	records := make([]ConfigRecord, len(l.records))
	copy(records, l.records)
	return records, nil
}
