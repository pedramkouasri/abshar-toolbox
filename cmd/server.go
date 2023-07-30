package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/pedramkousari/abshar-toolbox/cmd/patch"
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

		if (fileSrc == ""){
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("path is required"))
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
		w.Write([]byte("Good!"))
	})

	log.Println("Starting server on http://localhost:9990")
	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}


func fileExists(filename string) bool {
   info, err := os.Stat(filename)
   if os.IsNotExist(err) {
      return false
   }
   return !info.IsDir()
}