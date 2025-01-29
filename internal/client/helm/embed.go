package helm

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
)

//go:embed chart/*
var helmChartFS embed.FS

func GetHelmChart() (*chart.Chart, error) {
	var files []*loader.BufferedFile

	err := fs.WalkDir(helmChartFS, "chart", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			data, err := helmChartFS.ReadFile(path)
			if err != nil {
				return err
			}

			name := filepath.ToSlash(path) // Ensure the path is in a consistent format
			name = strings.TrimPrefix(name, "chart/")
			files = append(files, &loader.BufferedFile{Name: name, Data: data})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	chart, err := helmloader.LoadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("Error loading Helm chart: %w", err)
	}

	return chart, err
}
