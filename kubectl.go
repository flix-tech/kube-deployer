package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"os/exec"
	"strings"
)

func (client *KubeClient) GetDeployedBranchHashes(namespace string) ([]string, error) {
	usingContext := false

	if client.Context != "" {
		client.UseContext()
		usingContext = true
	}

	cmdName := "kubectl"
	cmdArgs := []string{
		"get",
		"deployments,rc,rs,pvc,svc,cronjobs",
		"-o",
		"jsonpath={.items[*].metadata.labels.branch_hash}",
		fmt.Sprintf("--namespace=%s", namespace),
	}

	if !usingContext {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--server=%s", client.Server))
		cmdArgs = append(cmdArgs, fmt.Sprintf("--token=%s", client.Token))
	}

	cmd := exec.Command(cmdName, cmdArgs...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return nil, errors.New(string(stderr.String()))
	}

	output := out.String()
	branches := strings.Split(string(output), " ")
	branches = Unique(branches)
	branches = Filter(branches, func(v string) bool {
		return v != ""
	})

	return branches, nil
}

func (client *KubeClient) DeleteObjectsByBranch(branchHash string, namespace string) (string, error) {
	usingContext := false

	if client.Context != "" {
		client.UseContext()
		usingContext = true
	}

	cmdName := "kubectl"
	cmdArgs := []string{
		"delete",
		"deployments,rc,rs,pvc,svc,cronjobs",
		"-l",
		fmt.Sprintf("branch_hash=%s", branchHash),
		fmt.Sprintf("--namespace=%s", namespace),
	}

	if !usingContext {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--server=%s", client.Server))
		cmdArgs = append(cmdArgs, fmt.Sprintf("--token=%s", client.Token))
	}

	cmd := exec.Command(cmdName, cmdArgs...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return "", errors.New(string(stderr.String()))
	}

	output := out.String()

	return output, err
}

func (client *KubeClient) Version() {
	usingContext := false

	if client.Context != "" {
		client.UseContext()
		usingContext = true
	}

	cmdArgs := []string{
		"version",
	}

	if !usingContext {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--server=%s", client.Server))
		cmdArgs = append(cmdArgs, fmt.Sprintf("--token=%s", client.Token))
	}

	client.runCommand("kubectl", cmdArgs, "")
	fmt.Println()
}

func (client *KubeClient) UseContext() {
	cmdArgs := []string{
		"config",
		"use-context",
		client.Context,
	}

	client.runCommand("kubectl", cmdArgs, "")
}

func (client *KubeClient) Apply(definition string, namespace string, dryRun bool) {
	usingContext := false

	if client.Context != "" {
		client.UseContext()
		usingContext = true
	}

	cmdArgs := []string{
		"apply",
		"-f",
		"-",
		fmt.Sprintf("--namespace=%s", namespace),
	}

	if !usingContext {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--server=%s", client.Server))
		cmdArgs = append(cmdArgs, fmt.Sprintf("--token=%s", client.Token))
	}

	if dryRun {
		cmdArgs = append(cmdArgs, "--dry-run")
	}

	client.runCommand("kubectl", cmdArgs, definition)
}

func (client *KubeClient) runCommand(cmdName string, cmdArgs []string, input string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stderr = cmd.Stdout

	if client.Verbose {
		color.Yellow(input)
	}

	stdin, _ := cmd.StdinPipe()
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	stdOutScanner := bufio.NewScanner(stdoutPipe)
	go func() {
		for stdOutScanner.Scan() {
			color.Green("%s\n", stdOutScanner.Text())
		}
	}()

	stdErrScanner := bufio.NewScanner(stderrPipe)
	go func() {
		for stdErrScanner.Scan() {
			color.Red("%s\n", stdErrScanner.Text())
		}
	}()

	cmd.Start()

	io.WriteString(stdin, input+"\n")
	stdin.Close()

	err := cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type KubeClient struct {
	Server  string
	Token   string
	Context string
	Verbose bool
}
