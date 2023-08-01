package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pedramkousari/abshar-toolbox/cmd/patch"
	"github.com/pedramkousari/abshar-toolbox/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

type ResponseServer struct {
	IsCompleted bool   `json:"is_complated"`
	IsFailed    bool   `json:"is_failed"`
	MessageFail string `json:"message_fail"`
	Percent     string `json:"percent"`
	State       string `json:"state"`
}

func startServer(cmd *cobra.Command) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	http.HandleFunc("/patch", func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		version := queryParams.Get("version")

		w.Header().Set("Content-Type", "application/json")

		if version == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "version is required"
			}`))
			return
		}

		baadbaanDir := viper.GetString("patch.update.baadbaan.directory")
		fileSrc := baadbaanDir + "/storage/app/patches/" + version

		if !fileExists(fileSrc) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "file not exists"
			}`))
			return
		}

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

			err := json.NewEncoder(w).Encode(&ResponseServer{
				IsCompleted: false,
				IsFailed:    true,
				MessageFail: "patch id is required",
				Percent:     "0",
				State:       "",
			})

			if err != nil {
				panic(err)
			}

			return
		}

		store := db.NewBoltDB()

		// p := store.Get(fmt.Sprintf(db.Format, patchId, db.Processed))
		percent := store.Get(fmt.Sprintf(db.Format, patchId, db.Percent))
		fmt.Println(percent)
		isComplete := store.Get(fmt.Sprintf(db.Format, patchId, db.IsCompleted))
		isFailed := store.Get(fmt.Sprintf(db.Format, patchId, db.IsFailed))
		messageFail := store.Get(fmt.Sprintf(db.Format, patchId, db.MessageFail))
		state := store.Get(fmt.Sprintf(db.Format, patchId, db.State))

		if len(percent) == 0 {
			w.WriteHeader(http.StatusOK)
			res := &ResponseServer{
				IsCompleted: false,
				IsFailed:    false,
				MessageFail: "",
				Percent:     "0",
				State:       "Not Started",
			}
			if err := json.NewEncoder(w).Encode(res); err != nil {
				panic(err)
			}
			return
		}

		res := ResponseServer{
			IsCompleted: isComplete[0] == 1,
			IsFailed:    isFailed[0] == 1,
			MessageFail: string(messageFail),
			Percent:     string(percent),
			State:       string(state),
		}

		if err := json.NewEncoder(w).Encode(&res); err != nil {
			panic(err)
		}

		if isFailed[0] == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

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
