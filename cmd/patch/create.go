/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"log"

	"github.com/pedramkousari/abshar-toolbox/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Patch",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err:= service.CreatePackage(viper.GetString("patch.baadbaan_directory"),viper.GetString("patch.branch1"), viper.GetString("patch.branch2")).Run(); err != nil {
			log.Fatal(err)
		}
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
