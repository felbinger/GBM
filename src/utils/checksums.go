package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hash"
	"io"
	"os"
	"path/filepath"
)

type entry struct {
	checksum []byte
	basename string
}

func GenerateChecksums(path string, sums []string) {
	for _, checksum := range sums {
		var checksums []entry

		// walk through directory and add generated checksums to slice
		_ = filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
			if !fi.Mode().IsRegular() {
				return nil
			}
			f, _ := os.Open(file)
			var h hash.Hash

			// this switch could be improved, might work using reflections
			switch checksum {
			case "md5":
				h = md5.New()
				break
			case "sha1":
				h = sha1.New()
				break
			case "sha256":
				h = sha256.New()
				break
			case "sha512":
				h = sha512.New()
				break
			default:
				log.Debug(fmt.Sprintf("Checksum %s does not exist!", checksum))
				return nil
			}
			_, err = io.Copy(h, f)
			if err != nil {
				log.Debug("Unable to generate checksum!", err)
				return nil
			}
			checksums = append(checksums, entry{h.Sum(nil), filepath.Base(file)})
			return nil
		})

		// append generated checksum to checksum file (create the file if it does not exist)
		f, err := os.OpenFile(fmt.Sprintf("%s/%ssum.txt", path, checksum), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		for _, e := range checksums {
			if _, err := f.WriteString(fmt.Sprintf("%x \t %s\n", e.checksum, e.basename)); err != nil {
				log.Println(err)
			}
		}
	}
}
