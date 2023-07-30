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

	"github.com/pedramkousari/abshar-toolbox/helpers"
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

		// var wg sync.WaitGroup;
		for _, packagex := range diffPackages {
			// wg.Add(1)

			go func(pkg types.CreatePackageParams) {
				// defer wg.Done()

				directory := viper.GetString(fmt.Sprintf("patch.create.%v.directory", pkg.ServiceName))
				conf := helpers.LoadEnv(directory)

				ctx := context.WithValue(context.Background(), "serviceName", pkg.ServiceName)
				path, err := service.CreatePackage(directory, pkg.PackageName1, pkg.PackageName2, conf).Run(ctx)
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
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
