package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pedramkousari/abshar-toolbox/cmd/patch"
	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/spf13/cobra"
)

// patchCmd represents the patch command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		startServer(cmd)
	},
}

func startServer(cmd *cobra.Command) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	http.HandleFunc("/patch", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		fileSrc := queryParams.Get("path")

		w.Header().Set("Content-Type", "application/json")

		if fileSrc == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "path is required"
			}`))
			return
		}

		w.Write([]byte(fileSrc))

		// if !fileExists(fileSrc) {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte("\n file not found!"))
		// 	return
		// }

		patch.UpdateCommand(fileSrc)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
				"message": "GOOD"
			}`))
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		patchId := queryParams.Get("patch-id")

		w.Header().Set("Content-Type", "application/json")

		if patchId == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "path id is required"
			}`))

			return
		}

		store := db.NewBoltDB()
		defer store.Close()

		// p := store.Get(fmt.Sprintf(db.Format, patchId, db.Processed))
		process := store.Get(fmt.Sprintf(db.Format, patchId, db.Processed))
		is_complete := store.Get(fmt.Sprintf(db.Format, patchId, db.IsCompleted))
		hasError := store.Get(fmt.Sprintf(db.Format, patchId, db.HasError))
		// errorMessage := store.Get(fmt.Sprintf(db.Format, patchId, db.ErrorMessage))

		if hasError[0] == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{
				"message": "Not Completed"
			}`))
			return
		}

		if is_complete[0] == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"message": "Completed"
			}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{
			"message": "",
			"process": %v
		}`, process[0])))

	})

	log.Println("Starting server on http://localhost:9990")
	log.Fatal(http.ListenAndServe("0.0.0.0:9990", nil))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
