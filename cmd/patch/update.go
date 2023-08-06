/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/logger"
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

		explode := strings.Split(fileSrc, "/")
		version := explode[len(explode)-1]
		db.StoreInit(version)
		logger.Info("Started")

		if err := UpdateCommand(fileSrc); err != nil {
			logger.Error(err)
		}
	},
}

func UpdateCommand(fileSrc string) error {
	if err := os.Mkdir("./temp", 0755); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("create directory err: %s", err)
		}
	}

	logger.Info("Created Temp Directory")

	if err := helpers.DecryptFile([]byte(key), fileSrc, strings.TrimSuffix(fileSrc, ".enc")); err != nil {
		return fmt.Errorf("Decrypt File err: %s", err)
	}

	logger.Info("Decrypted File")

	if err := helpers.UntarGzip(strings.TrimSuffix(fileSrc, ".enc"), "./temp"); err != nil {
		return fmt.Errorf("UnZip File err: %s", err)
	}
	logger.Info("UnZiped File")

	packagePathFile := "./temp/package.json"

	if _, err := os.Stat(packagePathFile); err != nil {
		return fmt.Errorf("package.json is err: %s", err)
	}

	logger.Info("Exists package.json")

	file, err := os.Open(packagePathFile)
	if err != nil {
		return fmt.Errorf("open package.json is err: %s", err)
	}
	logger.Info("Opened package.json")

	pkg := []types.Packages{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pkg)
	if err != nil {
		return fmt.Errorf("decode package.json is err: %s", err)
	}

	logger.Info("Decode package.json")

	diffPackages := service.GetPackageDiff(pkg)

	var wg sync.WaitGroup
	errCh := make(chan error)

	for _, packagex := range diffPackages {
		wg.Add(1)

		go func(pkg types.CreatePackageParams) {
			defer wg.Done()

			directory := viper.GetString(fmt.Sprintf("patch.update.%v.directory", pkg.ServiceName))

			ctx := context.WithValue(context.Background(), "information", map[string]string{
				"version":     pkg.PackageName2,
				"serviceName": pkg.ServiceName,
			})

			conf := helpers.LoadEnv(directory)
			err := service.UpdatePackage(directory, conf).Run(ctx, loading(pkg.ServiceName, true))
			if err != nil {
				errCh <- err
			}

		}(packagex)

	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	concatErrors := ""
	for err := range errCh {
		concatErrors += err.Error()
	}

	if concatErrors != "" {
		return errors.New(concatErrors)
	}

	return nil
}

func init() {
	PatchCmd.AddCommand(updateCmd)
}
