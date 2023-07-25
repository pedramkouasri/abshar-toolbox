package service

import (
	"fmt"
	"log"

	"github.com/pedramkousari/abshar-toolbox/types"
)

type updatePackage struct {
	directory       string
	composerCommand types.ComposerCommand
	migrateCommand  types.MigrateCommand
}

func UpdatePackage(srcDirectory string, composerCommand types.ComposerCommand, migrateCommand types.MigrateCommand) *updatePackage {
	if srcDirectory == "" {
		log.Fatal("src directory not initialized")
	}

	return &updatePackage{
		directory:       srcDirectory,
		composerCommand: composerCommand,
		migrateCommand:  migrateCommand,
	}
}

func (cr *updatePackage) Run() error {
	fmt.Println("Started ...")

	fmt.Printf("%+v", cr)

	return nil
}
