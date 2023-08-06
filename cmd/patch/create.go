/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/logger"
	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mapProcess = map[string]int{}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:                   "create PATH_OF_PACKAGE.JSON",
	Short:                 "Create PATH_OF_PACKAGE.JSON",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		packagePathFile := args[0]

		if _, err := os.Stat(packagePathFile); err != nil {
			log.Fatal("File Not Exists is Path: %s", packagePathFile)
		}

		file, err := os.Open(packagePathFile)
		if err != nil {
			panic(err)
		}

		pkg := []types.Packages{}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&pkg)
		if err != nil {
			log.Fatal(err)
		}

		diffPackages := service.GetPackageDiff(pkg)

		ch := make(chan string, len(diffPackages))
		var pathes []string = []string{packagePathFile}

		// var wg sync.WaitGroup;
		for _, packagex := range diffPackages {
			// wg.Add(1)

			go func(pkg types.CreatePackageParams) {
				// defer wg.Done()

				directory := viper.GetString(fmt.Sprintf("patch.create.%v.directory", pkg.ServiceName))
				conf := helpers.LoadEnv(directory)

				ctx := context.WithValue(context.Background(), "serviceName", pkg.ServiceName)
				path, err := service.CreatePackage(directory, pkg.PackageName1, pkg.PackageName2, conf).Run(ctx, loading(pkg.ServiceName, false))
				if err != nil {
					log.Fatal(err)
				}

				newPath := strings.ReplaceAll(path, "patch", packagex.ServiceName)
				if err := os.Rename(path, newPath); err != nil {
					log.Fatal(err)
				}

				ch <- newPath

			}(packagex)

		}

		for i := 0; i < len(diffPackages); i++ {
			// select {
			// 	case path := <-ch:
			// 		pathes = append(pathes, path)
			// }

			pathes = append(pathes, <-ch)
		}

		err = os.Mkdir("./builds", 0755)
		if err != nil {
			if os.IsNotExist(err) {
				panic(err)
			}
		}

		outputFile := fmt.Sprintf("./builds/%s.tar.gz", pkg[len(pkg)-1].Version)

		if err := helpers.TarGz(pathes, outputFile); err != nil {
			log.Fatal(err)
		}

		if err := helpers.EncryptFile([]byte(key), outputFile, outputFile+".enc"); err != nil {
			log.Fatal(err)
		}

		if err := os.Remove(outputFile); err != nil {
			fmt.Println("Error deleting file:", err)
			return
		}

		fmt.Println("\nCompleted :)")

		// wg.Wait()
	},
}

func init() {
	PatchCmd.AddCommand(createCmd)
}

func loading(serviceName string, store bool) func(types.Process) {
	mapProcess[serviceName] = 0

	return func(p types.Process) {
		mapProcess[serviceName] = p.State

		printProcess()

		if store {
			process := calcPercent()
			db.StorePercent(fmt.Sprint(process))

			if p.Message != "" {
				db.StoreInfo(p.Message)
				logger.Info(fmt.Sprintf("percent and message is %d:%s", process, p.Message))
			}
		}
	}
}

func printProcess() {
	cmdy := exec.Command("clear") //Linux example, its tested
	cmdy.Stdout = os.Stdout
	cmdy.Run()

	for serviceName, state := range mapProcess {
		fmt.Print(serviceName, ":[")
		for j := 0; j <= state; j++ {
			fmt.Print("=")
		}
		for j := state + 1; j <= 100; j++ {
			fmt.Print(" ")
		}
		fmt.Print("] %", state)
		fmt.Println()
	}
}

func calcPercent() int {
	sum := 0
	cnt := 0
	for _, state := range mapProcess {
		cnt++
		sum += state
	}

	if cnt == 0 {
		return 0
	}

	return int(sum / cnt)
}
