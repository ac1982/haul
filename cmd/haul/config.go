package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/storage"
	"github.com/spf13/pflag"
)

// applyConfig fills every option the command line did not set from config.json: keys are the option names in
// camelCase, bilibili's under "bilibili". An explicit --config must exist; the default one is optional.
func applyConfig(fs *pflag.FlagSet, defs []flagDef) error {
	path, _ := fs.GetString("config")
	explicit := path != ""
	if !explicit {
		path = storage.Path(storage.ConfigFile)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if explicit || !os.IsNotExist(err) {
			return errs.NewInput("Cannot read the config file %s: %v", path, err)
		}
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return errs.NewInput("Cannot read the config file %s: %v", path, err)
	}
	values := map[string]any{}
	flatten("", raw, values)
	known := map[string]flagDef{}
	for _, d := range defs {
		if d.config != "-" {
			known[d.configKey()] = d
		}
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		d, ok := known[key]
		if !ok {
			console.Warn(fmt.Sprintf("Unknown key %q in %s", key, console.PrettyPath(path)))
			continue
		}
		if fs.Changed(d.name) {
			continue
		}
		if err := fs.Set(d.name, configValue(values[key])); err != nil {
			return errs.NewInput("Bad value for %q in %s: %v", key, path, err)
		}
	}
	console.Status("Config  " + console.PrettyPath(path))
	return nil
}

func flatten(prefix string, in map[string]any, out map[string]any) {
	for k, v := range in {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if m, ok := v.(map[string]any); ok {
			flatten(key, m, out)
			continue
		}
		out[key] = v
	}
}

func configValue(v any) string {
	switch x := v.(type) {
	case []any:
		parts := make([]string, len(x))
		for i, p := range x {
			parts[i] = configValue(p)
		}
		return strings.Join(parts, ",")
	case float64:
		return fmt.Sprint(int64(x))
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}
