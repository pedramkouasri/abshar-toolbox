package helpers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func UntarGzip(sourceFile, destinationDir string) error {
	// Open the source gzip file
	gzipFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract each file from the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		if err != nil {
			return err
		}

		// Determine the file path for extraction
		target := filepath.Join(destinationDir, header.Name)

		// Create directories if necessary
		if header.Typeflag == tar.TypeDir {
			err := os.MkdirAll(target, 0755)
			if err != nil {
				return err
			}
			continue
		}

		// Create the file for extraction
		file, err := os.Create(target)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file data from the tar entry to the created file
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}