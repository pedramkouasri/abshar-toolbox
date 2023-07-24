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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type createPackage struct {
	directory       string
	branch1         string
	branch2         string
	composerCommand string
}

var currentDirectory string
var tempDir string

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

func CreatePackage(srcDirectory string, branch1 string, branch2 string, composerCommand string) *createPackage {

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
		directory:       srcDirectory,
		branch1:         branch1,
		branch2:         branch2,
		composerCommand: composerCommand,
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

func (cr *createPackage) Run() error {
	composerInstall(cr.composerCommand)

	if err := cr.fetch(); err != nil {
		return err
	}
	fmt.Printf("Fetch Completed \n")

	if err := cr.getDiffComposer(); err != nil {
		return err
	}
	fmt.Printf("Generated Diff.txt \n")

	// err := os.Chdir(cr.directory)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	createTarFile(cr.directory)
	fmt.Printf("Created Tar File \n")

	if composerChanged() {
		fmt.Printf("Composer Is Change \n")

		if cr.composerCommand != "" {
			cr.switchBranch()
			fmt.Printf("Branch Swiched  \n")

			fmt.Printf("Composer Install \n")
			composerInstall(cr.composerCommand)
			fmt.Printf("Composer Installed \n")
		}

		// cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && composer-lock-diff  --from %s  --to %s --json --pretty --only-prod > %s/composer-lock-diff.json", cr.directory, cr.branch1, cr.branch2, tempDir))

		// _, err := cmd.Output()
		// if err != nil {
		// 	return err
		// }

		addDiffPackageToTarFile(cr.directory)
		fmt.Printf("Added Diff Packages To Tar File \n")

	}

	copyTarFileToTempDirectory(cr.directory)
	fmt.Printf("Moved Tar File \n")

	gzipTarFile()
	fmt.Printf("GZiped Tar File \n")

	return nil
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
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter", "ACMR", cr.branch1, cr.branch2)

	cmd.Dir = cr.directory

	res, err := cmd.Output()
	if err != nil {
		return err
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
func composerInstall(_ string) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Define the container ID or name
	containerID := "baadbaan_php_pedram"

	// Define the command and arguments to execute
	cmd := []string{"composer", "update", "--no-scripts"}

	// Prepare the exec create options
	createOptions := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	// Create the exec instance
	execCreateResp, err := cli.ContainerExecCreate(context.Background(), containerID, createOptions)
	if err != nil {
		panic(err)
	}

	// Attach to the exec instance
	execAttachResp, err := cli.ContainerExecAttach(context.Background(), execCreateResp.ID, types.ExecStartCheck{})
	if err != nil {
		panic(err)
	}
	defer execAttachResp.Close()

	// Read the output from the exec instance
	output := make([]byte, 4096)
	_, err = execAttachResp.Reader.Read(output)
	if err != nil {
		panic(err)
	}

	// Print the output
	fmt.Println(string(output))
	log.Fatal("END")
}

// func composerInstall(composerCommand string) {
// 	// cmd := exec.Command("sh", "-c", composerCommand)

// 	cmd := exec.Command("/bin/sh", "-c", "docker exec -ti baadbaan_php_pedram ls -la")
// 	res, err := cmd.Output()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	fmt.Println(string(res))

// 	cmd = exec.Command("docker", "exec", "-ti", "baadbaan_php_pedram", "composer", "install", "--no-scripts", "--no-interaction")
// 	fmt.Println("docker", "exec", "-ti", "baadbaan_php_pedram", "composer", "install", "--no-scripts", "--no-interaction")
// 	res, err = cmd.Output()
// 	if err != nil {
// 		fmt.Println(string(res))
// 		log.Fatal(err.Error())
// 	}
// }

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

	// 	file := bytes.NewBufferString(`{
	// 	"changes": {
	// 		"abshar/acunetix-baadbaan": ["fd3197e", "REMOVED", ""],
	// 		"abshar/goals-relation": ["49bcb86", "REMOVED", ""],
	// 		"abshar/process_management": ["5804c6d", "REMOVED", ""],
	// 		"aws/aws-crt-php": ["v1.2.1", "REMOVED", ""],
	// 		"aws/aws-sdk-php": [
	// 		"3.275.5",
	// 		"3.171.19",
	// 		"https://github.com/aws/aws-sdk-php/compare/3.275.5...3.171.19"
	// 		],
	// 		"vlucas/phpdotenv": [
	// 		"v2.6.9",
	// 		"v2.6.6",
	// 		"https://github.com/vlucas/phpdotenv/compare/v2.6.9...v2.6.6"
	// 		]
	// 	}
	// 	}
	// `)

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
