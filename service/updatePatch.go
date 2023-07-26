package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pedramkousari/abshar-toolbox/types"
)

var current_time = time.Now()

const backaupSqlDir = "./backaupSql"

type updatePackage struct {
	directory       string
	composerCommand types.ComposerCommand
	migrateCommand  types.MigrateCommand
}

func UpdatePackage(srcDirectory string, composerCommand types.ComposerCommand, migrateCommand types.MigrateCommand) *updatePackage {
	if srcDirectory == "" {
		log.Fatal("src directory not initialized")
	}

	return &updatePackage{
		directory:       srcDirectory,
		composerCommand: composerCommand,
		migrateCommand:  migrateCommand,
	}
}

func (cr *updatePackage) Run() error {
	fmt.Println("Started ...")

	fmt.Printf("%+v", cr)

	return nil
}

func changePermision(dir string) {
	if _,err := exec.Command("chown", fmt.Sprintf("-R %s.%s %s", "www-data", "www-data", dir)).Output();err != nil {
		panic(err)
	}
}

func backupFileWithGit(dir string, version string){
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git diff")
	cmd.Dir = dir
	cmd.Stdout = stdOut

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	if len(output) != 0 {
		createBranch(dir, version)
		gitAdd(dir, version)
		gitCommit(dir, version)
	}
}

func createBranch(dir string, version string) error {
	cmd := exec.Command("git",fmt.Sprintf("checkout -b patch-before-update-%s-%s", version, current_time))
	cmd.Dir = dir
	if _, err := cmd.Output();err != nil {
		return err
	}
	return nil
}


func gitAdd(dir string, version string) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if _, err := cmd.Output();err != nil {
		panic(err)
	}
}

func gitCommit(dir string, version string) {

	exportCmd := exec.Command("export", "HOME=/tmp")
	exportCmd.Stdout = os.Stdout
	exportCmd.Stderr = os.Stderr
	err := exportCmd.Run()
	if err != nil {
		fmt.Println("Error exporting command:", err)
		return
	}

	cmd := exec.Command("git", "config", "--global", "user.email", "persianped@gmail.com")
	if _, err := cmd.Output();err != nil {
		panic(err)
	}

	cmd = exec.Command("git", "config", "--global", "user.name", "pedram kousari")
	if _, err := cmd.Output();err != nil {
		panic(err)
	}


	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("backup befor update patch %s time: %s", version, current_time))
	cmd.Dir = dir
	if _, err := cmd.Output();err != nil {
		panic(err)
	}
}

func backupDatabase(mc types.Command, directory string, serviceName string) {
	err := os.Mkdir(backaupSqlDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("The directory named", backaupSqlDir, "exists")
		} else {
			log.Fatalln(err)
		}
	}


	sqlFileName := fmt.Sprintf("%s-%s.sql", serviceName, current_time)
	file, err := os.Create(backaupSqlDir + "/" + sqlFileName)
	if err != nil {
		panic(err)
	}


	var command []string
	if mc.Type == types.DockerCommandType {
		command = strings.Fields(fmt.Sprintf("docker exec %s %s", mc.Container, mc.Cmd))
	} else {
		command = strings.Fields(mc.Cmd)
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	cmd.Dir = directory

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}


// func runConfig(ct types.CommandType, dir string){
// 	var command []string
// 	if ct.Type == types.DockerCommandType {
// 		command = strings.Fields(fmt.Sprintf("docker exec %s %s", mc.Container, mc.Cmd))
// 	} else {
// 		command = strings.Fields(mc.Cmd)
// 	}

// 	cmd := exec.Command(command[0], command[1:]...)
// 	cmd.Stdout = file
// 	cmd.Stderr = os.Stderr
// 	cmd.Dir = directory

// 	err = cmd.Run()
// 	if err != nil {
// 		panic(err)
// 	}
// }
