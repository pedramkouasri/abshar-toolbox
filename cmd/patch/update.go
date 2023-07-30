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
	"sync"

	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:                   "update PATH",
	Short:                 "Update PATH",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		fileSrc := args[0]

		if _, err := os.Stat(fileSrc); err != nil {
			log.Fatal("File Not Exists is Path: %s", fileSrc)
		}

		UpdateCommand(fileSrc)
	},
}

func UpdateCommand(fileSrc string) {
	if err := os.Mkdir("./temp", 0755); err != nil {
		if os.IsNotExist(err) {
			panic(err)
		}
	}

	if err := helpers.DecryptFile([]byte(key), fileSrc, strings.TrimSuffix(fileSrc, ".enc")); err != nil {
		panic(err)
	}

	if err := helpers.UntarGzip(strings.TrimSuffix(fileSrc, ".enc"), "./temp"); err != nil {
		panic(err)
	}

	packagePathFile := "./temp/package.json"

	if _, err := os.Stat(packagePathFile); err != nil {
		panic(err)
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

	var wg sync.WaitGroup
	for _, packagex := range diffPackages {
		wg.Add(1)

		go func(pkg types.CreatePackageParams) {
			defer wg.Done()

			directory := viper.GetString(fmt.Sprintf("patch.update.%v.directory", pkg.ServiceName))

			ctx := context.WithValue(context.Background(), "information", map[string]string{
				"version":     packagex.PackageName2,
				"serviceName": packagex.ServiceName,
			})

			conf := helpers.LoadEnv(directory)
			err := service.UpdatePackage(directory, conf).Run(ctx)
			if err != nil {
				log.Fatal(err)
			}

		}(packagex)

	}

	wg.Wait()

	fmt.Println("\nCompleted :)")
}

func init() {
	PatchCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
