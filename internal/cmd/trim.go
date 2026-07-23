package cmd

import (
	"fmt"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// trimConfigCmd removes fields from a single --config file whenever the
// value already matches the code-level default or the type's zero value —
// the mechanical counterpart to diff-config's "redundant" section, using the
// same ground truth (the marshaled provider.ProvideDefaultConfig() struct,
// i.e. exactly what export-default-config prints). This deliberately
// includes fields with no explicit SetDefault(...) call, e.g. an unfilled
// secret sitting at its zero value: the user is responsible for setting
// config values, and an omitted field behaves identically to one written out
// as empty (validate:"required" tags still fail startup either way) — the
// only difference is visibility in this one file, and diff-config's "missing
// from file" section (same ground truth) still surfaces anything trimmed
// away as available to set, so nothing silently disappears from view.
var trimConfigCmd = &cobra.Command{
	Use:   "trim-config",
	Short: "remove fields from a config file that are redundant with the code-level defaults",
	Long:  `prints the given --config file with every field that exactly matches the code-level default removed, plus any map left empty as a result`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTrimConfig()
	},
}

func init() {
	rootCmd.AddCommand(trimConfigCmd)
}

func runTrimConfig() error {
	if len(configFilenames) != 1 {
		return fmt.Errorf("trim-config takes exactly one --config file, got %d", len(configFilenames))
	}
	path := configFilenames[0]

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", path, err)
	}
	var tree yaml.MapSlice
	if err := yaml.Unmarshal(b, &tree); err != nil {
		return fmt.Errorf("unable to parse %s: %w", path, err)
	}

	defaultYAML, err := yaml.Marshal(provider.ProvideDefaultConfig())
	if err != nil {
		return fmt.Errorf("unable to marshal default config: %w", err)
	}
	var defaultTree map[interface{}]interface{}
	if err := yaml.Unmarshal(defaultYAML, &defaultTree); err != nil {
		return fmt.Errorf("unable to unmarshal default config: %w", err)
	}
	defaultValues := flattenSettings(defaultTree)

	trimmed, removed := trimRedundant("", tree, defaultValues)

	out, err := yaml.Marshal(trimmed)
	if err != nil {
		return fmt.Errorf("unable to marshal trimmed config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "removed %d field(s) redundant with the code-level defaults\n", removed)
	fmt.Print(string(out))
	return nil
}

// trimRedundant walks a yaml.MapSlice tree and drops any leaf whose value
// matches the default at that dotted path. A map left empty after trimming
// its children is dropped too, so a fully-redundant section (e.g. one where
// every field happens to match the default) disappears entirely rather than
// being left behind as `{}`. Returns the trimmed node (nil if it became
// empty) and how many leaves were removed.
func trimRedundant(prefix string, node interface{}, defaults map[string]string) (interface{}, int) {
	if m, ok := node.(yaml.MapSlice); ok {
		var out yaml.MapSlice
		removed := 0
		for _, item := range m {
			key := fmt.Sprintf("%v", item.Key)
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			trimmedVal, r := trimRedundant(path, item.Value, defaults)
			removed += r
			if trimmedVal == nil {
				continue
			}
			out = append(out, yaml.MapItem{Key: item.Key, Value: trimmedVal})
		}
		if len(out) == 0 {
			return nil, removed
		}
		return out, removed
	}

	dv, ok := defaults[prefix]
	if !ok {
		return node, 0
	}
	if valuesEqual(fmt.Sprintf("%v", node), dv) {
		return nil, 1
	}
	return node, 0
}
