/*
Copyright © 2023 pedram kousari <pedrsianped@gmail.com>
*/
package main

import (
	"log"
	"os"

	"github.com/pedramkousari/abshar-toolbox/cmd"
	"github.com/pedramkousari/abshar-toolbox/db"
)

func main() {
	store := db.NewBoltDB()
	defer store.Close()

	logFile, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	cmd.Execute()
}
