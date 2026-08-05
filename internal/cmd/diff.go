package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// diffConfigCmd compares whatever --config file(s) were given against
// export-default-config's own output, field by field, and reports which
// values are genuine overrides, which are redundant (already covered by the
// default), which don't exist in the Config struct's schema at all, and
// which have a code-level default but are entirely absent from the file
// (e.g. a whole map-typed section like storage.file_stores that nothing in
// the file ever populated).
//
// Ground truth is the *marshaled* provider.ProvideDefaultConfig() struct —
// the same thing export-default-config prints — not provider.SetDefaultConfig
// applied to a bare Viper instance. Those two differ: Go's YAML marshaling
// emits every non-omitempty field's zero value (e.g. `password: ""`) whether
// or not that field ever got an explicit v.SetDefault(...) call, so comparing
// against the Viper registry alone flagged plenty of fields — anything with
// no registered default but a zero value in both the file and the struct —
// as "only in file" even though they came straight from export-default-config
// with no manual edit at all.
var diffConfigCmd = &cobra.Command{
	Use:     "diff-config",
	Short:   "show drift between --config file(s) and the code-level defaults",
	Long:    `flags every field in the given --config file(s) as either overriding a code-level default, redundant with it, having no default at all, or being a code-level default the file never set`,
	PreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiffConfig()
	},
}

func init() {
	rootCmd.AddCommand(diffConfigCmd)
}

func runDiffConfig() error {
	// viper.AllSettings() lowercases every key internally, which loses the
	// real casing for anything with an uppercase letter (e.g. a camelCase
	// field like accessControlAllowOrigins). Join file/default values on a
	// lowercased key throughout, but keep the original-case string for
	// display -- preferring whichever source has the real casing: the raw
	// file parse, then (authoritative when the key is part of the schema)
	// the default struct's own marshaled tags.
	display := map[string]string{}

	fileValues := map[string]string{}
	for k, v := range flattenSettings(viper.AllSettings()) {
		lk := strings.ToLower(k)
		fileValues[lk] = v
		display[lk] = k
	}

	// viper.AllSettings() also drops a key entirely when its YAML value is
	// an explicit null, making "set to null" indistinguishable from "never
	// set". Backfill from the raw file(s) directly, so an explicit null
	// compares correctly against the default instead of looking entirely
	// unset -- this also fixes the casing for every key actually in a file.
	rawValues, err := flattenRawConfigFiles(configFilenames)
	if err != nil {
		return err
	}
	for k, v := range rawValues {
		lk := strings.ToLower(k)
		if _, ok := fileValues[lk]; !ok {
			fileValues[lk] = v
		}
		display[lk] = k
	}

	defaultYAML, err := yaml.Marshal(provider.ProvideDefaultConfig())
	if err != nil {
		return fmt.Errorf("unable to marshal default config: %w", err)
	}
	var defaultTree map[interface{}]interface{}
	if err := yaml.Unmarshal(defaultYAML, &defaultTree); err != nil {
		return fmt.Errorf("unable to unmarshal default config: %w", err)
	}
	defaultValues := map[string]string{}
	for k, v := range flattenSettings(defaultTree) {
		lk := strings.ToLower(k)
		defaultValues[lk] = v
		display[lk] = k // the struct's own tag casing is authoritative
	}

	seen := map[string]bool{}
	var keys []string
	for k := range fileValues {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	for k := range defaultValues {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	sort.Strings(keys)

	fmt.Println("=== overrides (differ from the code-level default) ===")
	for _, k := range keys {
		fv, fOk := fileValues[k]
		dv, dOk := defaultValues[k]
		if fOk && dOk && !valuesEqual(fv, dv) {
			fmt.Printf("%s\n  file:    %s\n  default: %s\n", display[k], fv, dv)
		}
	}

	fmt.Println("\n=== redundant (same as the code-level default — candidate to trim) ===")
	for _, k := range keys {
		fv, fOk := fileValues[k]
		dv, dOk := defaultValues[k]
		if fOk && dOk && valuesEqual(fv, dv) {
			fmt.Printf("%s = %s\n", display[k], fv)
		}
	}

	fmt.Println("\n=== only in file (no code-level default exists) ===")
	for _, k := range keys {
		fv, fOk := fileValues[k]
		_, dOk := defaultValues[k]
		if fOk && !dOk {
			fmt.Printf("%s = %s\n", display[k], fv)
		}
	}

	fmt.Println("\n=== missing from file (has a code-level default, not set here) ===")
	for _, k := range keys {
		_, fOk := fileValues[k]
		dv, dOk := defaultValues[k]
		// viper.AllSettings() drops a key entirely when its YAML value is an
		// explicit null, making "set to null" and "never set" indistinguishable
		// on the file side — so a nil default can never be reliably compared.
		if !fOk && dOk && dv != "<nil>" {
			fmt.Printf("%s = %s\n", display[k], dv)
		}
	}

	return nil
}

// flattenSettings turns a nested settings tree — either viper.AllSettings()'s
// map[string]interface{} or yaml.v2's map[interface{}]interface{} — into
// dot-path -> stringified-value pairs, so two configs can be compared key by
// key regardless of which decoder produced them.
func flattenSettings(v interface{}) map[string]string {
	out := map[string]string{}
	flattenSettingsInto("", v, out)
	return out
}

// valuesEqual compares two stringified settings values, treating equal
// time.Duration values as equal even when their string forms differ (e.g.
// "72h" from a YAML file vs. "72h0m0s" from a native time.Duration default).
func valuesEqual(a, b string) bool {
	if a == b {
		return true
	}
	da, errA := time.ParseDuration(a)
	db, errB := time.ParseDuration(b)
	return errA == nil && errB == nil && da == db
}

// flattenRawConfigFiles reads and merges the given YAML files directly (same
// order/precedence as initConfig(): later files win on duplicate keys), then
// flattens the result. Unlike viper.AllSettings(), this preserves explicit
// YAML nulls as present keys with a nil value.
func flattenRawConfigFiles(paths []string) (map[string]string, error) {
	merged := map[interface{}]interface{}{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("unable to read %s: %w", p, err)
		}
		var tree map[interface{}]interface{}
		if err := yaml.Unmarshal(b, &tree); err != nil {
			return nil, fmt.Errorf("unable to parse %s: %w", p, err)
		}
		deepMergeYAML(merged, tree)
	}
	return flattenSettings(merged), nil
}

// deepMergeYAML merges src into dst in place: nested maps are merged
// recursively, everything else in src overwrites the value in dst.
func deepMergeYAML(dst, src map[interface{}]interface{}) {
	for k, v := range src {
		if srcMap, ok := v.(map[interface{}]interface{}); ok {
			if dstMap, ok := dst[k].(map[interface{}]interface{}); ok {
				deepMergeYAML(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func flattenSettingsInto(prefix string, v interface{}, out map[string]string) {
	if m, ok := v.(map[string]interface{}); ok {
		for k, vv := range m {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			flattenSettingsInto(p, vv, out)
		}
		return
	}
	if m, ok := v.(map[interface{}]interface{}); ok {
		for k, vv := range m {
			p := fmt.Sprintf("%v", k)
			if prefix != "" {
				p = prefix + "." + p
			}
			flattenSettingsInto(p, vv, out)
		}
		return
	}
	out[prefix] = fmt.Sprintf("%v", v)
}
