package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/viper"
)

var current_time = time.Now()

const (
	backaupSqlDir = "./backaupSql"
)

type updatePackage struct {
	directory string
	config    *helpers.ConfigService
}

func UpdatePackage(srcDirectory string, cnf *helpers.ConfigService) *updatePackage {
	return &updatePackage{
		directory: srcDirectory,
		config:    cnf,
	}
}

func (cr *updatePackage) Run(ctx context.Context, progress func(types.Process)) error {
	if cr.directory == "" {
		return fmt.Errorf("src directory not initialized")
	}

	information := ctx.Value("information").(map[string]string)
	version := information["version"]

	if err := changePermision(cr.directory); err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}
	progress(types.Process{
		State:   1,
		Message: "Changed Permission",
	})

	if err := backupFileWithGit(cr.directory, version); err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}
	progress(types.Process{
		State:   2,
		Message: "Backup File Complete With git",
	})

	if err := backupDatabase(cr.directory, information["serviceName"], cr.config); err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}
	progress(types.Process{
		State:   4,
		Message: "Backup Database Complete",
	})

	if err := extractTarFile(information["serviceName"], cr.directory); err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}
	progress(types.Process{
		State:   6,
		Message: "Extracted Tar File",
	})

	if err := composerDumpAutoload(cr.directory, cr.config); err != nil {
		return fmt.Errorf("Composer Dump Autoload Failed Error Is: %s", err)
	}
	progress(types.Process{
		State:   8,
		Message: "Composer Dup Autoload complete",
	})

	if err := migrateDB(cr.directory, cr.config); err != nil {
		return fmt.Errorf("Migrate Database Failed Error Is: %s", err)
	}

	progress(types.Process{
		State:   10,
		Message: "Migrated Database",
	})

	return nil
}

func changePermision(dir string) error {
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
		return fmt.Errorf("Error retrieving information for user %s: %s\n", username, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("Error retrieving convert uid to string for uid %s: %s\n", u.Uid, err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("Error retrieving convert gid to string for gid %s: %s\n", u.Gid, err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if err = os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("Failed to change ownership of %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk to filepath error in err  %s\n", err)
	}

	return nil
}

func backupFileWithGit(dir string, version string) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = dir
	cmd.Stdout = stdOut

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Git diff Failed Error is : %v\n", err)
	}

	if stdOut.Len() > 0 {
		if err := createBranch(dir, version); err != nil {
			return fmt.Errorf("Create Branch Failed Error is : %v\n", err)
		}
		if err := gitAdd(dir, version); err != nil {
			return fmt.Errorf("Git Add Failed Error is : %v\n", err)
		}
		if err := gitCommit(dir, version); err != nil {
			return fmt.Errorf("Git Commit Failed Error is : %v\n", err)
		}
	}

	return nil
}

func createBranch(dir string, version string) error {
	cmd := exec.Command("git", strings.Fields(fmt.Sprintf("checkout -b patch-before-update-%s-%d", version, current_time.Unix()))...)

	cmd.Stdout = nil

	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func gitAdd(dir string, version string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

func gitCommit(dir string, version string) error {

	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		return fmt.Errorf("Error setting environment variable: %v", err)
	}

	cmd := exec.Command("git", "config", "--global", "user.email", "persianped@gmail.com")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git set config Email Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "config", "--global", "user.name", "pedram kousari")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git set config UserName Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "config", "--global", "--add", "safe.directory", dir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git Set Safe Directory Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("backup befor update patch %s time: %d", version, current_time.Unix()))
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
	}

	return nil
}

func backupDatabase(dir string, serviceName string, cnf *helpers.ConfigService) error {
	err := os.Mkdir(backaupSqlDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			// fmt.Println("The directory named", backaupSqlDir, "exists")
		} else {
			return fmt.Errorf("Create backaupSql Directory Failed error is: %v", err)
		}
	}

	sqlFileName := fmt.Sprintf("%s-%d.sql", serviceName, current_time.Unix())
	file, err := os.Create(backaupSqlDir + "/" + sqlFileName)
	if err != nil {
		return fmt.Errorf("Create sql fileFailed error is: %v", err)
	}
	defer file.Close()

	host, _ := cnf.Get("DB_HOST")
	port, _ := cnf.Get("DB_PORT")
	datbase, _ := cnf.Get("DB_DATABASE")
	username, _ := cnf.Get("DB_USERNAME")
	password, _ := cnf.Get("DB_PASSWORD")
	sqlCommand := fmt.Sprintf(sqlDumpCommand, username, password, host, port, datbase)

	commandType := getCommandType(cnf)

	var command []string
	if commandType == types.DockerCommandType {
		composeDir := viper.GetString("patch.update.docker-compose-directory") + "/docker-compose.yaml"
		command = strings.Fields(fmt.Sprintf(`docker compose -f %s run --rm %s %s`, composeDir, host, sqlCommand))
	} else {
		command = strings.Fields(sqlCommand)
	}

	cmd := exec.Command(command[0], command[1:]...)
	// cmd := exec.Command("sh", "-c", strings.Join(command, " "))

	cmd.Stdout = file
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Dump sql Failed error is: %v", err)
	}

	return nil
}

func extractTarFile(serviceName string, dir string) error {
	cmd := exec.Command("tar", "-zxf", "./temp/"+serviceName+".tar.gz", "-C", dir)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func composerDumpAutoload(dir string, cnf *helpers.ConfigService) error {

	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		return fmt.Errorf("Error setting environment variable : %v", err)
	}

	var command []string = getCommand(composerDumpCommand, cnf)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func migrateDB(dir string, cnf *helpers.ConfigService) error {
	var command []string = getCommand(migrateCommand, cnf)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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
