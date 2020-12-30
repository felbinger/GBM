package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func Cp(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		log.Debug(fmt.Sprintf("%s is not a regular file", src))
		return 0, nil
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dest.Close()

	nBytes, err := io.Copy(dest, source)
	if err != nil {
		return 0, err
	}
	return nBytes, err
}