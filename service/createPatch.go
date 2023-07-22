package service

import (
	"fmt"
	"os/exec"
)

type createPackage struct {
	directory string
	branch1 string
	branch2 string
}

func CreatePackage(srcDirectory string, branch1 string, branch2 string) *createPackage {
	return &createPackage{
		directory: srcDirectory,
		branch1: branch1,
		branch2: branch2,
	}
}

func (cr *createPackage) Run() error {
	return cr.getDiffComposer()
}

func (cr *createPackage) getDiffComposer() error {
	// git diff --name-only --diff-filter=ACMR {lastTag} {current_tag} > diff.txt'
	cmd := exec.Command("sh", "-c", fmt.Sprintf("git --git-dir %s/.git  diff --name-only --diff-filter ACMR %s %s > %s/diff.txt", cr.directory, cr.branch1, cr.branch2, cr.directory))

	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}