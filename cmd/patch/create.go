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
	"strings"

	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/logger"
	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

		serviceCount := len(diffPackages)

		// var wg sync.WaitGroup;
		for _, packagex := range diffPackages {
			// wg.Add(1)

			go func(pkg types.CreatePackageParams) {
				// defer wg.Done()

				directory := viper.GetString(fmt.Sprintf("patch.create.%v.directory", pkg.ServiceName))
				conf := helpers.LoadEnv(directory)

				ctx := context.WithValue(context.Background(), "serviceName", pkg.ServiceName)
				path, err := service.CreatePackage(directory, pkg.PackageName1, pkg.PackageName2, conf).Run(ctx, loading(pkg.ServiceName, serviceCount, false))
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

func loading(serviceName string, serviceCount int, store bool) func(types.Process) {
	process := 0

	return func(p types.Process) {
		fmt.Print("\r", serviceName, ":[")
		for j := 0; j <= p.State; j++ {
			fmt.Print("=")
		}
		for j := p.State + 1; j <= 10; j++ {
			fmt.Print(" ")
		}
		fmt.Print("] ", p.State*10, "%")

		process += p.State * (10 / serviceCount)
		if store {
			db.StorePercent(fmt.Sprint(process))
			if p.Message != "" {
				db.StoreInfo(p.Message)
				logger.Info(fmt.Sprintf("percent and message is %d:%s", process, p.Message))
			}
		}
	}
}
