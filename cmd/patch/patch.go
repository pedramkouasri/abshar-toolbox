/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package patch

import (
	"github.com/spf13/cobra"
)

const key = "e10adc3949ba59abbe56e057f20f883e"

// patchCmd represents the patch command
var PatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Create Or Update Patch",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// patchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// patchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
