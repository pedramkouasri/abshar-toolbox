package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pedramkousari/abshar-toolbox/types"
)

type createPackage struct {
	directory       string
	branch1         string
	branch2         string
	composerCommand types.ComposerCommand
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

func CreatePackage(srcDirectory string, branch1 string, branch2 string, composerCommand types.ComposerCommand) *createPackage {

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

func (cr * createPackage) GenerateDiffJson () {

		file, err := os.Create(tempDir+"/composer-lock-diff.json");
		if err!=nil {
			panic(err)
		}
		
		cmd := exec.Command("composer-lock-diff", "--from", "remotes/origin/"+cr.branch1,  "--to" ,"remotes/origin/"+cr.branch2, "--json", "--pretty", "--only-prod")
		cmd.Stdout = file
		// cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = cr.directory

		err = cmd.Run()
		if err != nil {
			panic(err)
		}
}

func (cr *createPackage) Run() (string,error) {
	fmt.Println("Started ...")

	if err := cr.fetch(); err != nil {
		return "" , err
	}
	fmt.Printf("Fetch Completed \n")

	if err := cr.getDiffComposer(); err != nil {
		return "", err
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

		if cr.composerCommand.Cmd != "" {
			cr.switchBranch()
			fmt.Printf("Branch Swiched  \n")

			composerInstall(cr.composerCommand, cr.directory)
			fmt.Printf("Composer Installed \n")
		}

		
		cr.GenerateDiffJson()
		fmt.Printf("Generated Diff Package Composer \n")

		addDiffPackageToTarFile(cr.directory)
		fmt.Printf("Added Diff Packages To Tar File \n")

	}

	copyTarFileToTempDirectory(cr.directory)
	fmt.Printf("Moved Tar File \n")

	gzipTarFile()
	fmt.Printf("GZiped Tar File \n")

	return tempDir+"/patch.tar.gz", nil
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

func composerInstall(cc types.ComposerCommand, directory string) {
	var command []string
	if cc.Type == types.DockerCommandType {
		command = strings.Fields(fmt.Sprintf("docker exec %s %s",cc.Container,cc.Cmd))
	}else{
		command = strings.Fields(cc.Cmd)
	}

	cmd := exec.Command(command[0], command[1:]...)
	// cmd.Stdout = nil
	// cmd.Stderr = os.Stderr
	cmd.Dir = directory

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

// func composerInstall(_ string) {
// 	cli, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Define the container ID or name
// 	containerID := "baadbaan_php_pedram"

// 	// Define the command and arguments to execute
// 	cmd := []string{"composer", "update", "--no-scripts"}

// 	// Prepare the exec create options
// 	createOptions := types.ExecConfig{
// 		Cmd:          cmd,
// 		AttachStdout: true,
// 		AttachStderr: true,
// 		Tty:          false,
// 	}

// 	// Create the exec instance
// 	execCreateResp, err := cli.ContainerExecCreate(context.Background(), containerID, createOptions)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Attach to the exec instance
// 	execAttachResp, err := cli.ContainerExecAttach(context.Background(), execCreateResp.ID, types.ExecStartCheck{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer execAttachResp.Close()

// 	// Read the output from the exec instance
// 	output := make([]byte, 4096)
// 	_, err = execAttachResp.Reader.Read(output)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Print the output
// 	fmt.Println(string(output))
// 	log.Fatal("END")
// }

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
