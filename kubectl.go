package main

import (
	"fmt"
	"os/exec"
	"bufio"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
	"bytes"
	"errors"
)

func (client *KubeClient) GetDeployedBranchHashes(namespace string) ([]string, error) {
	cmdName := "kubectl"
	cmdArgs := []string{
		"get",
		"deployments,rc,rs,pvc,svc,cronjobs",
		"-o",
		"jsonpath={.items[*].metadata.labels.branch_hash}",
		fmt.Sprintf("--server=%s", client.Server),
		fmt.Sprintf("--token=%s", client.Token),
		fmt.Sprintf("--namespace=%s", namespace),
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
	cmdName := "kubectl"
	cmdArgs := []string{
		"delete",
		"deployments,rc,rs,pvc,svc,cronjobs",
		"-l",
		fmt.Sprintf("branch_hash=%s", branchHash),
		fmt.Sprintf("--namespace=%s", namespace),
		fmt.Sprintf("--token=%s", client.Token),
		fmt.Sprintf("--server=%s", client.Server),
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

func (client *KubeClient) DeleteCronJob(cronJob string, namespace string) {
	cmdName := "kubectl"
	cmdArgs := []string{
		"delete",
		"cronJob",
		fmt.Sprintf("--namespace=%s", namespace),
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Output()
}

func (client *KubeClient) Version() {
	client.runCommand("kubectl", []string{
		"version",
		fmt.Sprintf("--server=%s", client.Server),
		fmt.Sprintf("--token=%s", client.Token),
	}, "")
	fmt.Println()
}

func (client *KubeClient) Apply(definition string, namespace string, dryRun bool) {
	cmdArgs := []string{
		"apply",
		"-f",
		"-",
		fmt.Sprintf("--server=%s", client.Server),
		fmt.Sprintf("--token=%s", client.Token),
		fmt.Sprintf("--namespace=%s", namespace),
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
	Verbose bool
}
