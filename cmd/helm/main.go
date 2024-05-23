// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"intelligence/pkg/utils/semver"
)

const (
	helmCommand           = "helm"
	helmVersionSubcommand = "version"
	helmLintSubcommand    = "lint"
	helmInstallSubcommand = "install"

	helmMinimumVersion = "v3.14.2"
	helmMaximumVersion = "v4.0.0"

	chartsDir      = "charts"
	configFileName = "config.json"
	valuesPrefix   = "values"
	valuesEnvSep   = "-"
	valuesSuffix   = ".yaml"
)

var (
	allowedEnvs = map[string]struct{}{
		"local": {},
		"dev":   {},
		"prod":  {},
	}
)

// helmConfiguration is a json file in each Helm chart that represents options for the Helm CLI commands.
type helmConfiguration struct {
	ChartName string
	Order     int    `json:"order"`
	Namespace string `json:"namespace"`
}

// ensureHelmIsInstalled checks the system PATH to see if it can find the helm binary.
func ensureHelmIsInstalled() {
	_, err := exec.LookPath(helmCommand)
	if err != nil {
		log.Fatal("Helm is not installed.")
	}
}

// ensureHelmVersionIsSufficient compares the helm's version is within the accepted range.
func ensureHelmVersionIsSufficient() {
	cmd := exec.Command(helmCommand, helmVersionSubcommand, "--short")
	log.Println(cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error getting helm version.")
	}

	helmCurrentVersion := strings.TrimSpace(string(output))
	if helmCurrentVersion == "" {
		log.Fatal("Could not find the helm version.")
	}

	minimumVersionCheck, err := semver.Compare(helmCurrentVersion, helmMinimumVersion)
	if err != nil {
		log.Fatalf("Could not compare Helm versions (%s).", err.Error())
	}
	if minimumVersionCheck < 0 {
		log.Fatalf("Helm version %s is too old (>= %s).", helmCurrentVersion, helmMinimumVersion)
	}

	maximumVersionCheck, err := semver.Compare(helmCurrentVersion, helmMaximumVersion)
	if err != nil {
		log.Fatalf("Could not compare Helm versions (%s).", err.Error())
	}
	if maximumVersionCheck >= 0 {
		log.Fatalf("Helm version %s is too new (>= %s).", helmCurrentVersion, helmMaximumVersion)
	}

	log.Printf("Helm version is within the range (%s <= %s < %s).\n", helmMinimumVersion, helmCurrentVersion, helmMaximumVersion)
}

// ensureFileExists logs an error and exists if the file path does not exist.
func ensureFileExists(path string) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Path %s does not exist.", path)
		} else {
			log.Fatalf("Error checking if %s exists.", path)
		}
	}
}

// getHelmChartConfigs returns each charts configuration.
func getHelmChartConfigs() []*helmConfiguration {
	ensureFileExists(chartsDir)

	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		log.Fatal("Error reading helm charts directory.")
	}

	configs := make([]*helmConfiguration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			relativePath := filepath.Join(chartsDir, entry.Name(), configFileName)

			valuesFileName := valuesPrefix + valuesSuffix
			valuesPath := filepath.Join(chartsDir, entry.Name(), valuesFileName)
			ensureFileExists(valuesPath)

			for envName := range allowedEnvs {
				overrideFileName := valuesPrefix + valuesEnvSep + envName + valuesSuffix
				overrideValuesPath := filepath.Join(chartsDir, entry.Name(), overrideFileName)
				ensureFileExists(overrideValuesPath)
			}

			data, err := os.ReadFile(relativePath)
			if err != nil {
				log.Fatalf("Error reading helm configuration file for chart %s.", relativePath)
			}

			chartConfig := &helmConfiguration{
				ChartName: entry.Name(),
			}

			jsonDecoder := json.NewDecoder(bytes.NewBuffer(data))
			jsonDecoder.DisallowUnknownFields()

			if err := jsonDecoder.Decode(chartConfig); err != nil {
				log.Fatalf("Error parsing configuration file for chart %s (%s).", entry.Name(), err.Error())
			}

			configs = append(configs, chartConfig)
		}
	}

	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Order < configs[j].Order
	})

	uniqueOrder := make(map[int]string)
	for _, config := range configs {
		if duplicate, ok := uniqueOrder[config.Order]; ok {
			log.Fatalf("Duplicate order for chart %s and %s.", duplicate, config.ChartName)
		}
		uniqueOrder[config.Order] = config.ChartName
	}

	for _, config := range configs {
		log.Printf("Found chart %s with order %d.\n", config.ChartName, config.Order)
	}

	return configs
}

// helmLintCharts walks the charts directory and lints each chart.
func helmLintCharts(configs []*helmConfiguration) {
	for _, config := range configs {
		log.Printf("Linting chart %s.\n", config.ChartName)

		chartPath := filepath.Join(chartsDir, config.ChartName)
		cmd := exec.Command(helmCommand, helmLintSubcommand, "--with-subcharts", "--quiet", "--strict", chartPath)
		log.Println(cmd.String())

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Fatal("Error when invoking helm lint.")
		}

		if err := cmd.Wait(); err != nil {
			log.Fatal("Error when running helm lint.")
		}
	}
}

// helmUpdateDependencies updates the dependencies of each helm chart.
func helmUpdateDependencies(configs []*helmConfiguration) {
	for _, config := range configs {
		log.Printf("Updating dependencies for chart %s.\n", config.ChartName)

		chartPath := filepath.Join(chartsDir, config.ChartName)
		cmd := exec.Command(helmCommand, "dependency", "update", chartPath)
		log.Println(cmd.String())

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Fatal("Error when invoking helm dependency update.")
		}

		if err := cmd.Wait(); err != nil {
			log.Fatal("Error when running helm dependency update.")
		}
	}
}

// helmInstallCharts installs each chart in order.
func helmInstallCharts(configs []*helmConfiguration, environment string) {
	for _, config := range configs {
		log.Printf("Installing chart %s.\n", config.ChartName)

		chartPath := filepath.Join(chartsDir, config.ChartName)
		valuesFileName := valuesPrefix + valuesSuffix
		valuesPath := filepath.Join(chartsDir, config.ChartName, valuesFileName)
		overrideFileName := valuesPrefix + valuesEnvSep + environment + valuesSuffix
		overrideValuesPath := filepath.Join(chartsDir, config.ChartName, overrideFileName)

		options := []string{
			"upgrade",
			config.ChartName,
			chartPath,
			"--values", valuesPath,
			"--values", overrideValuesPath,
			"--install",
			"--atomic",
			"--cleanup-on-fail",
			"--wait",
			"--timeout", "30m0s",
			"--qps", "5",
			"--history-max", "3",
		}

		if config.Namespace != "" {
			options = append(options, "--namespace", config.Namespace)
			options = append(options, "--create-namespace")
		}

		cmd := exec.Command(helmCommand, options...)
		log.Println(cmd.String())

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Fatal("Error when invoking helm upgrade.")
		}

		if err := cmd.Wait(); err != nil {
			log.Fatal("Error when running helm upgrade.")
		}
	}
}

func main() {
	const minimumNumberOfArgs = 2
	if len(os.Args) < minimumNumberOfArgs {
		log.Fatal("helm command requires at least one argument.")
	}

	ensureHelmIsInstalled()
	ensureHelmVersionIsSufficient()

	configs := getHelmChartConfigs()

	switch strings.TrimSpace(strings.ToLower(os.Args[1])) {
	case helmLintSubcommand:
		const requiredArgCountForLint = 2
		if len(os.Args) != requiredArgCountForLint {
			log.Fatal("Helm lint has no arguments.")
		}
		helmLintCharts(configs)
	case helmInstallSubcommand:
		const requiredArgCountForInstall = 3
		if len(os.Args) != requiredArgCountForInstall {
			log.Fatal("Helm install requires an environment argument.")
		}

		environment := os.Args[2]
		if _, envIsAllowed := allowedEnvs[environment]; !envIsAllowed {
			log.Fatalf("Environment %s is not supported.", environment)
		}

		helmUpdateDependencies(configs)
		helmInstallCharts(configs, environment)
	default:
		log.Fatalf("Unknown command '%s'.", os.Args[1])
	}
}
