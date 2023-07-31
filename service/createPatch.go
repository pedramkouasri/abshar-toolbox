package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/types"
)

type createPackage struct {
	directory string
	branch1   string
	branch2   string
	config    *helpers.ConfigService
}

var currentDirectory string
var tempDir string

var excludePath = []string{".env", "vmanager.json"}

func init() {
	currentDirectory, _ = os.Getwd()

	os.RemoveAll(currentDirectory + "/temp")

	err := os.Mkdir(currentDirectory+"/temp", 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("The directory named", currentDirectory+"/temp", "exists")
		} else {
			log.Fatalln(err)
		}
	}
}

func CreatePackage(srcDirectory string, branch1 string, branch2 string, cnf *helpers.ConfigService) *createPackage {

	if srcDirectory == "" {
		log.Fatal("src directory not initialized")
	}

	if branch1 == "" {
		log.Fatal("branch 1 not initialized")
	}

	if branch2 == "" {
		log.Fatal("branch 2 not initialized")
	}

	createTempDirectory(srcDirectory)

	return &createPackage{
		directory: srcDirectory,
		branch1:   branch1,
		branch2:   branch2,
		config:    cnf,
	}
}

func (cr *createPackage) switchBranch() {
	cmd := exec.Command("git", "checkout", cr.branch2)
	cmd.Dir = cr.directory
	_, err := cmd.Output()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (cr *createPackage) GenerateDiffJson() {

	file, err := os.Create(tempDir + "/composer-lock-diff.json")
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("composer-lock-diff", "--from", "remotes/origin/"+cr.branch1, "--to", "remotes/origin/"+cr.branch2, "--json", "--pretty", "--only-prod")
	cmd.Stdout = file
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cr.directory

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func (cr *createPackage) Run(ctx context.Context, progress func(state int)) (string, error) {
	// fmt.Println("Started ...")

	progress(0)

	if err := cr.fetch(); err != nil {
		return "", err
	}

	progress(1)

	if err := cr.getDiffComposer(); err != nil {
		return "", err
	}
	progress(2)
	// fmt.Printf("Generated Diff.txt \n")

	// err := os.Chdir(cr.directory)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	createTarFile(cr.directory)
	progress(3)
	// fmt.Printf("Created Tar File \n")

	if composerChanged() {
		// fmt.Printf("Composer Is Change \n")

		progress(4)

		cr.switchBranch()
		// fmt.Printf("Branch Swiched  \n")
		progress(5)

		composerInstall(cr.directory, cr.config)
		// fmt.Printf("Composer Installed \n")

		progress(6)
		cr.GenerateDiffJson()
		// fmt.Printf("Generated Diff Package Composer \n")

		progress(7)
		addDiffPackageToTarFile(cr.directory)
		// fmt.Printf("Added Diff Packages To Tar File \n")

	}

	progress(8)

	copyTarFileToTempDirectory(cr.directory)
	// fmt.Printf("Moved Tar File \n")
	progress(9)

	gzipTarFile()
	// fmt.Printf("GZiped Tar File \n")
	progress(10)

	return tempDir + "/patch.tar.gz", nil
}

func (cr *createPackage) fetch() error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("git --git-dir %s/.git  fetch", cr.directory))

	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func (cr *createPackage) getDiffComposer() error {
	// git diff --name-only --diff-filter=ACMR {lastTag} {current_tag} > diff.txt'
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter", "ACMR", "remotes/origin/"+cr.branch1, "remotes/origin/"+cr.branch2)

	cmd.Dir = cr.directory

	res, err := cmd.Output()
	if err != nil {
		return err
	}

	for _, path := range excludePath {
		res = []byte(strings.ReplaceAll(string(res), path, ""))
	}

	ioutil.WriteFile(tempDir+"/diff.txt", res, 0666)
	return nil
}

func createTarFile(directory string) {
	// tar -cf patch.tar --files-from=diff.txt
	cmd := exec.Command("tar", "-cf", "./patch.tar", fmt.Sprintf("--files-from=%s/diff.txt", tempDir))

	cmd.Dir = directory

	if _, err := cmd.Output(); err != nil {
		if err.Error() != "exit status 2" {
			log.Fatal(err)
		}
	}
}

func gzipTarFile() {
	// cd {baadbaan_path} && gzip -f patch.tar
	cmd := exec.Command("gzip", "-f", fmt.Sprintf("%s/patch.tar", tempDir))

	_, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
}

func composerChanged() bool {
	diffFile, err := os.Open(tempDir + "/diff.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer diffFile.Close()

	scanner := bufio.NewScanner(diffFile)

	var exists bool = false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "composer.lock" {
			exists = true
			break
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return exists
}

func composerInstall(directory string, cnf *helpers.ConfigService) {
	safeCommand := getCommand(gitSafeDirectory, cnf)
	cmd := exec.Command(safeCommand[0], safeCommand[1:]...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	cmd.Dir = directory

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	command := getCommand(composerInstallCommand, cnf)
	cmd = exec.Command(command[0], command[1:]...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	cmd.Dir = directory

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func addDiffPackageToTarFile(directory string) {
	for packageName := range getDiffPackages() {
		cmd := exec.Command("tar", "-rf", "./patch.tar", "vendor/"+packageName)
		cmd.Dir = directory
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func getDiffPackages() map[string][]string {
	//TODO::remove
	file, err := os.Open(tempDir + "/composer-lock-diff.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	type ChangesType struct {
		Changes map[string][]string `json:"changes"`
	}

	changesInstance := ChangesType{}

	if err := json.NewDecoder(file).Decode(&changesInstance); err != nil {
		log.Fatal(err)
	}

	for index, packageName := range changesInstance.Changes {
		if packageName[1] == "REMOVED" {
			delete(changesInstance.Changes, index)
		}
	}

	return changesInstance.Changes
}

func copyTarFileToTempDirectory(directory string) {
	if err := os.Rename(directory+"/patch.tar", tempDir+"/patch.tar"); err != nil {
		log.Fatal(err.Error())
	}
}

func createTempDirectory(directory string) {
	splitDir := strings.Split(directory, "/")
	tempDir = currentDirectory + "/temp/" + splitDir[len(splitDir)-1]

	err := os.Mkdir(tempDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("The directory named", tempDir, "exists")
		} else {
			log.Fatalln(err)
		}
	}
}

func getCommand(cmd string, cnf *helpers.ConfigService) []string {
	commandType := getCommandType(cnf)

	var command []string
	if commandType == types.DockerCommandType {
		containerName, _ := cnf.Get("CONTAINER_NAME")
		command = strings.Fields(fmt.Sprintf("docker exec %s %s", containerName, cmd))
	} else {
		command = strings.Fields(cmd)
	}

	return command
}

func getCommandType(cnf *helpers.ConfigService) types.CommandType {
	containerName, _ := cnf.Get("CONTAINER_NAME")
	commandType := types.ShellCommandType

	if containerName != "" {
		commandType = types.DockerCommandType
	}

	return commandType
}
