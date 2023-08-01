/*
Copyright © 2023 pedram kousari <persianped@gmail.com>
*/
package patch

import (
	"fmt"

	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var infoCmd = &cobra.Command{
	Use:                   "info PATH ID",
	Short:                 "info PATH ID",
	Long:                  ``,
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		information(args[0])
	},
}

func information(patchId string) {
	store := db.NewBoltDB()
	defer store.Close()

	// p := store.Get(fmt.Sprintf(db.Format, patchId, db.Processed))
	process := store.Get(fmt.Sprintf(db.Format, patchId, db.State))
	is_complete := store.Get(fmt.Sprintf(db.Format, patchId, db.IsCompleted))
	hasError := store.Get(fmt.Sprintf(db.Format, patchId, db.IsFailed))
	errorMessage := store.Get(fmt.Sprintf(db.Format, patchId, db.MessageFail))

	fmt.Printf("%v - %v - %v - %v", process[0], is_complete[0], hasError[0], errorMessage)
}

func init() {
	PatchCmd.AddCommand(infoCmd)
}
