/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/pedramkousari/abshar-toolbox/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	packagePathFile string = "./package.json"
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
		
		var wg sync.WaitGroup;
		for _, packagex := range diffPackages {
			wg.Add(1)

			go func (pkg types.CreatePackageParams) {
				defer wg.Done()

				
				directory := viper.GetString(fmt.Sprintf("patch.create.%v.directory",pkg.ServiceName))
				var cc types.ComposerCommand;
				if err :=  viper.UnmarshalKey(fmt.Sprintf("patch.create.%v.composer_command",pkg.ServiceName), &cc); err != nil{
					panic(err)
				}
				
				if err:= service.CreatePackage(directory,pkg.PackageName1,pkg.PackageName2, cc).Run(); err != nil {
					log.Fatal(err)
				}
			}(packagex)

		}
		wg.Wait()
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
