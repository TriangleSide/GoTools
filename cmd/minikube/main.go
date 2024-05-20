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
	"strings"

	"intelligence/pkg/utils/semver"
)

const (
	minikubeCommand           = "minikube"
	minikubeVersionSubcommand = "version"
	minikubeDeleteSubcommand  = "delete"
	minikubeStartSubcommand   = "start"
	minikubeStatusSubcommand  = "status"

	minikubeProfileFlag = "--profile=intelligence"

	minikubeMinimumVersion = "v1.33.0"
	minikubeMaximumVersion = "v2.0.0"
)

// ensureMinikubeIsInstalled checks the system PATH to see if it can find the minikube binary.
func ensureMinikubeIsInstalled() {
	_, err := exec.LookPath(minikubeCommand)
	if err != nil {
		log.Fatal("Minikube is not installed.")
	}
}

// ensureMinikubeVersionIsSufficient compares the minikube's version command output with the defined minimum version.
func ensureMinikubeVersionIsSufficient() {
	cmd := exec.Command(minikubeCommand, minikubeVersionSubcommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error getting minikube version.")
	}

	minikubeCurrentVersion := ""
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "minikube version:") {
			versionParts := strings.Fields(line)
			minikubeCurrentVersion = versionParts[len(versionParts)-1]
			break
		}
	}
	if minikubeCurrentVersion == "" {
		log.Fatal("Could find the Minikube version.")
	}

	minimumVersionCheck, err := semver.Compare(minikubeCurrentVersion, minikubeMinimumVersion)
	if err != nil {
		log.Fatalf("Could not compare minikube versions (%s).", err.Error())
	}
	if minimumVersionCheck < 0 {
		log.Fatalf("Minikube version %s is too old (>= %s).", minikubeCurrentVersion, minikubeMinimumVersion)
	}

	maximumVersionCheck, err := semver.Compare(minikubeCurrentVersion, minikubeMaximumVersion)
	if err != nil {
		log.Fatalf("Could not compare minikube versions (%s).", err.Error())
	}
	if maximumVersionCheck >= 0 {
		log.Fatalf("Minikube version %s is too new (>= %s).", minikubeCurrentVersion, minikubeMaximumVersion)
	}

	log.Printf("Minikube version is accepted (%s <= %s < %s).\n", minikubeMinimumVersion, minikubeCurrentVersion, minikubeMaximumVersion)
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
	case minikubeStartSubcommand:
		if minikubeTestClusterExists() {
			log.Fatal("The minikube test cluster is already running.")
		}
		minikubeCreateTestCluster()
	default:
		log.Fatalf("Unknown command '%s'.", os.Args[1])
	}
}
