/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pedramkousari/abshar-toolbox/helpers"
	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	packagePathFile string = "./package.json"
	key = "e10adc3949ba59abbe56e057f20f883e"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Patch",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if _,err := os.Stat(packagePathFile); err != nil {
			panic(err)
		}
		

		// if err:=helpers.EncryptFile([]byte("1111111111111111"), "./builds/archive.tar.gz", "./builds/archive.tar.gz.enc"); err!= nil {
		// 	panic(err)
		// }

		// if err:=helpers.DecryptFile([]byte("1111111111111111"), "./builds/archive.tar.gz.enc", "./builds/archive1.tar.gz"); err!= nil {
		// 	panic(err)
		// }

		// if err:= helpers.UntarGzip("./builds/archive1.tar.gz", "./builds"); err != nil {
		// 	panic(err)
		// }

		// os.Exit(1)

		file, err := os.Open(packagePathFile)
		if err!= nil {
			panic(err)	
		}

		pkg := []types.Packages{}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&pkg)
		if err!= nil {
			log.Fatal(err)			
		}
		
		diffPackages := service.GetPackageDiff(pkg)

		ch := make(chan string, len(diffPackages))
		var pathes []string;
		
		// var wg sync.WaitGroup;
		for _, packagex := range diffPackages {
			// wg.Add(1)

			go func (pkg types.CreatePackageParams) {
				// defer wg.Done()

				
				directory := viper.GetString(fmt.Sprintf("patch.create.%v.directory",pkg.ServiceName))
				var cc types.ComposerCommand;
				if err :=  viper.UnmarshalKey(fmt.Sprintf("patch.create.%v.composer_command",pkg.ServiceName), &cc); err != nil{
					panic(err)
				}
				
				path, err:= service.CreatePackage(directory,pkg.PackageName1,pkg.PackageName2, cc).Run()
				if err != nil {
					log.Fatal(err)
				}

				ch <- path


			}(packagex)

		}


		for i:= 0; i< len(diffPackages); i++{
			// select {
			// 	case path := <-ch:
			// 		pathes = append(pathes, path)
			// }

			pathes = append(pathes, <-ch)
		}

		fmt.Println(pathes)

		err = os.Mkdir("./builds", 0755)
		if err != nil {
			if os.IsNotExist(err) {
				panic(err)
			}
		}

		outputFile := fmt.Sprintf("./builds/%s.tar.gz",pkg[len(pkg) - 1].Version)

		if err := helpers.TarGz(pathes, outputFile); err != nil {
			log.Fatal(err)
		}

		if err:=helpers.EncryptFile([]byte(key), outputFile, outputFile+".enc"); err!= nil {
			log.Fatal(err)
		}

		
		if err := os.Remove(outputFile); err != nil {
			fmt.Println("Error deleting file:", err)
			return
		}

		fmt.Println("Completed :)")

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
