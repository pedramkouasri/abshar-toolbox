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
				path, err := service.CreatePackage(directory, pkg.PackageName1, pkg.PackageName2, conf).Run(ctx, loading(pkg.ServiceName, serviceCount, nil))
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

func loading(serviceName string, serviceCount int, store *db.BoltDB) func(state int) {
	process := 0

	patchId := "1"

	if store != nil {
		store.Set(fmt.Sprintf(db.Format, patchId, db.IsCompleted), []byte{0})
		store.Set(fmt.Sprintf(db.Format, patchId, db.IsFailed), []byte{0})
		store.Set(fmt.Sprintf(db.Format, patchId, db.MessageFail), []byte{})
		store.Set(fmt.Sprintf(db.Format, patchId, db.Percent), []byte(string(process)))
		store.Set(fmt.Sprintf(db.Format, patchId, db.State), []byte{})
	}

	return func(state int) {
		fmt.Print("\r", serviceName, ":[")
		for j := 0; j <= state; j++ {
			fmt.Print("=")
		}
		for j := state + 1; j <= 10; j++ {
			fmt.Print(" ")
		}
		fmt.Print("] ", state*10, "%")

		if store != nil {
			process += state * (10 / serviceCount)
			store.Set(fmt.Sprintf(db.Format, patchId, db.Percent), []byte(string(process)))
		}
	}
}
