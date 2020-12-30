package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateTarball(path string, content []string, ignore []string) error {
	usingCreatedFile := false

	file, err := os.Create(path)
	if err != nil {
		log.Debug(fmt.Sprintf("Unable to create archive file %s.", path))
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, contentPath := range content {
		// check if content path exist
		info, infoErr := os.Stat(contentPath)
		if info != nil && !info.Mode().IsDir() || infoErr != nil {
			continue
		}
		err = filepath.Walk(contentPath, func(file string, fi os.FileInfo, err error) error {

			// do not add paths to tar file that should be ignored
			for _, ignorePath := range ignore {
				if file == ignorePath || strings.HasPrefix(file, ignorePath+"/") {
					return nil
				}
			}

			// return on any error
			if err != nil {
				return err
			}

			// return on non-regular files
			if !fi.Mode().IsRegular() {
				return nil
			}

			// create a new dir/file header
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}

			// update the name to correctly reflect the desired destination when untaring
			header.Name = strings.TrimPrefix(strings.Replace(file, contentPath, filepath.Base(contentPath), -1), string(filepath.Separator))

			// write the header
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// open files for taring
			f, err := os.Open(file)
			if err != nil {
				return err
			}

			// copy file data into tar writer
			if _, err := io.Copy(tarWriter, f); err != nil {
				return err
			}

			usingCreatedFile = true

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()

			return nil
		})
		if err != nil {
			return err
		}
	}

	// remove tar file if there is nothing in it..
	if !usingCreatedFile {
		log.Debug(fmt.Sprintf("The Archive %s has been deleted, cause there where nothing to put into it.", path))
		err := os.RemoveAll(path)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}
