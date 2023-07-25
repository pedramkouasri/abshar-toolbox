package helpers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func TarGz(files []string, outputFile string) error {
	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}


	// Create a gzip writer
	gw := gzip.NewWriter(outFile)	
	defer gw.Close()

	// Create a tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over the input files
	for _, file := range files {
		err = addFileToTar(file, tw)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToTar(file string, tw *tar.Writer) (error) {
	// Open the input file
	inFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Get the file information
	info, err := inFile.Stat()
	if err != nil {
		return err
	}

	// Create a tar header based on the file info
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	// Set the name of the file within the tar archive
	header.Name = filepath.Base(file)

	// Write the header the tar writer
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy the file content to the tar writer
	_, err = io.Copy(tw, inFile)
	if err != nil {
		return err
	}

	return nil
}