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
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	minikubeCommand           = "minikube"
	minikubeVersionSubcommand = "version"
	minikubeDeleteSubcommand  = "delete"
	minikubeStartSubcommand   = "start"
	minikubeStatusSubcommand  = "status"

	minikubeProfileFlag = "--profile=intelligence"

	minikubeVersionRegex   = "^v[0-9]+[.][0-9]+[.][0-9]+$"
	minikubeMinimumVersion = "v1.33.0"
)

// ensureMinikubeIsInstalled checks the system PATH to see if it can find the minikube binary
// and verifies that the file is executable.
func ensureMinikubeIsInstalled() {
	path, err := exec.LookPath(minikubeCommand)
	if err != nil {
		log.Fatal("Minikube is not installed.")
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal("Failed to get the Minikube file status.")
	}

	if fileInfo.Mode()&0100 == 0 {
		log.Fatal("Minikube is not executable.")
	}
}

// ensureMinikubeVersionIsSufficient compares the minikube's version command output with the defined minimum version.
func ensureMinikubeVersionIsSufficient() {
	cmd := exec.Command(minikubeCommand, minikubeVersionSubcommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error getting minikube version.")
	}

	versionRegex, err := regexp.Compile(minikubeVersionRegex)
	if err != nil {
		log.Fatal("Error compiling regex for minikube version.")
	}

	minikubeCurrentVersion := ""
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "minikube version:") {
			versionParts := strings.Fields(line)
			minikubeCurrentVersion = versionParts[len(versionParts)-1]

			if !versionRegex.MatchString(minikubeCurrentVersion) {
				log.Fatal("Could not parse Minikube version.")
			}

			break
		}
	}

	if minikubeCurrentVersion == "" {
		log.Fatal("Could find the Minikube version.")
	}

	currentVersionParts := strings.Split(minikubeCurrentVersion[1:], ".")
	minimumVersionParts := strings.Split(minikubeMinimumVersion[1:], ".")

	if len(currentVersionParts) != len(minimumVersionParts) {
		log.Fatal("Mismatch in the formatting of the minimum version and current version.")
	}

	for i := 0; i < len(currentVersionParts); i++ {
		currentVersionPart, err := strconv.Atoi(currentVersionParts[i])
		if err != nil {
			log.Fatal("Could not parse the current minikube version.")
		}

		minimumVersionPart, err := strconv.Atoi(minimumVersionParts[i])
		if err != nil {
			log.Fatal("Could not parse the minimum minikube version.")
		}

		if currentVersionPart < minimumVersionPart {
			log.Fatal("Minikube version " + minikubeCurrentVersion + " is not sufficient (< " + minikubeMinimumVersion + ").")
		}
	}

	log.Println("Minikube version " + minikubeCurrentVersion + " is sufficient (>= " + minikubeMinimumVersion + ").")
}

// minikubeTestClusterExists runs the `minikube status` command to check if the cluster exists.
func minikubeTestClusterExists() bool {
	cmd := exec.Command(minikubeCommand, minikubeStatusSubcommand, minikubeProfileFlag)

	if err := cmd.Start(); err != nil {
		log.Fatal("Error when invoking minikube status.")
	}

	if err := cmd.Wait(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return false
		}
		log.Fatal("Error when running minikube status.")
	}

	return true
}

// minikubeDeleteTestCluster deletes the test cluster created by minikubeCreateTestCluster.
func minikubeDeleteTestCluster() {
	cmd := exec.Command(minikubeCommand, minikubeDeleteSubcommand, minikubeProfileFlag)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("Error when invoking minikube delete.")
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal("Error when running minikube delete.")
	}
}

// minikubeDeleteTestCluster creates a test cluster for running the system locally. It uses sensible defaults.
func minikubeCreateTestCluster() {
	flags := []string{
		minikubeStartSubcommand,
		minikubeProfileFlag,
		"--driver=docker",
		"--memory=2g",
		"--cpus=2",
		"--interactive=false",
		"--nodes=3",
		"--cni=false",
		"--network-plugin=cni",
		"--extra-config=kubeadm.pod-network-cidr=192.168.0.0/16",
		"--subnet=172.16.0.0/24",
	}

	cmd := exec.Command(minikubeCommand, flags...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatal("Error when invoking minikube start.")
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal("Error when running minikube start.")
	}
}

// main runs helper commands for minikube binary.
func main() {
	const maxNumberOfArgs = 2
	if len(os.Args) != maxNumberOfArgs {
		log.Fatal("minikube command requires exactly one argument.")
	}

	ensureMinikubeIsInstalled()
	ensureMinikubeVersionIsSufficient()

	switch strings.TrimSpace(strings.ToLower(os.Args[1])) {
	case minikubeDeleteSubcommand:
		minikubeDeleteTestCluster()
	case minikubeStatusSubcommand:
		if minikubeTestClusterExists() {
			log.Fatal("The minikube test cluster is already running.")
		}
		minikubeCreateTestCluster()
	default:
		log.Fatalf("Unknown command '%s'.", os.Args[1])
	}
}
