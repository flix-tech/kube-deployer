package main

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func BranchHashes(projectDir string) map[string]string {
	cmdName := "git"
	cmdArgs := []string{
		"ls-remote",
		"--heads",
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = projectDir

	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	stringOutput := string(output)

	var rp = regexp.MustCompile(`(?m:\n)`)
	lines := rp.Split(stringOutput, -1)
	lines = Filter(lines, func(v string) bool {
		return v != ""
	})

	var branches = map[string]string{}

	for _, line := range lines {
		lineParts := strings.Split(line, "/")

		var branch string
		for _, branchSplit := range lineParts[2:] {
			branch = branch + branchSplit + "/"
		}

		branch = strings.TrimRight(branch, "/")
		branch = strings.TrimLeft(branch, " ")
		branch = strings.TrimRight(branch, " ")

		branchHash := MD5(branch)
		branches[branchHash] = branch
	}

	return branches
}
