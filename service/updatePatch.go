package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pedramkousari/abshar-toolbox/types"
)

var current_time = time.Now()

const backaupSqlDir = "./backaupSql"

type updatePackage struct {
	directory       string
	composerCommand types.Command
	migrateCommand  types.Command
	sqlDumpCommand  types.Command
}

func UpdatePackage(srcDirectory string, composerCommand types.Command, migrateCommand types.Command, sqlDumpC types.Command) *updatePackage {
	if srcDirectory == "" {
		log.Fatal("src directory not initialized")
	}

	return &updatePackage{
		directory:       srcDirectory,
		composerCommand: composerCommand,
		migrateCommand:  migrateCommand,
		sqlDumpCommand: sqlDumpC,
	}
}

func (cr *updatePackage) Run(ctx context.Context) error {
	information := ctx.Value("information").(map[string]string)
	version := information["version"]
	progress := loading(information["serviceName"])
	progress(0)

	changePermision(cr.directory)
	progress(1)

	backupFileWithGit(cr.directory, version)
	progress(2)

	backupDatabase(cr.sqlDumpCommand, cr.directory, information["serviceName"])
	progress(4)


	extractTarFile(information["serviceName"],cr.directory)
	progress(6)

	composerDumpAutoload(cr.composerCommand, cr.directory)
	progress(8)

	migrateDB(cr.composerCommand, cr.directory)
	progress(10)

	return nil
}

func changePermision(dir string) {
	// fmt.Println(fmt.Sprintf("-R %s.%s %s", "www-data", "www-data", dir))
	// cmd := exec.Command("chown", fmt.Sprintf("-R %s.%s %s", "www-data", "www-data", dir))

	// cmd.


	// if err != nil {
	// 	fmt.Println(res)
	// 	panic(err)
	// }
	username := "www-data"

	u, err := user.Lookup(username)
	if err != nil {
		fmt.Printf("Error retrieving information for user %s: %s\n", username, err)
		return
	}

	uid,err := strconv.Atoi(u.Uid)
	if err!=nil {
		panic(err)
	}

	gid,err := strconv.Atoi(u.Gid)
	if err!=nil {
		panic(err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if err = os.Chown(path, uid, gid); err != nil {
			fmt.Printf("Failed to change ownership of %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

}

func backupFileWithGit(dir string, version string){
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff")
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
			// fmt.Println("The directory named", backaupSqlDir, "exists")
		} else {
			log.Fatalln(err)
		}
	}


	sqlFileName := fmt.Sprintf("%s-%d.sql", serviceName, current_time.Unix())
	file, err := os.Create(backaupSqlDir + "/" + sqlFileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()


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

func extractTarFile(serviceName string, dir string){
	fmt.Println("tar -zxf ./temp/"+serviceName+".tar.gz -C "+dir)
	cmd := exec.Command("tar", "-zxf", "./temp/"+serviceName+".tar.gz", "-C", dir)
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run();err != nil {
		panic(err)
	}
}

func composerDumpAutoload(cc types.Command, dir string){

	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		fmt.Println("Error setting environment variable:", err)
		return
	}

	var command []string
	if cc.Type == types.DockerCommandType {
		command = strings.Fields(fmt.Sprintf("docker exec %s %s", cc.Container, cc.Cmd))
	} else {
		command = strings.Fields(cc.Cmd)
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func migrateDB(mc types.Command, dir string){

	var command []string
	if mc.Type == types.DockerCommandType {
		command = strings.Fields(fmt.Sprintf("docker exec %s %s", mc.Container, mc.Cmd))
	} else {
		command = strings.Fields(mc.Cmd)
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir

	
	if err := cmd.Run();err != nil {
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