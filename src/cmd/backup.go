package cmd

import (
	"fmt"
	"github.com/felbinger/GBM/utils"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Mode",
	Run: func(cmd *cobra.Command, args []string) {
		backup(Configuration)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func backup(conf utils.Config) {
	date := time.Now().Format("2006-01-02")
	err := os.MkdirAll(fmt.Sprintf("%s/%s", conf.Location, date), 0755)
	if err != nil {
		log.Fatal("Unable to create backup directory: %s", conf.Location)
		return
	}

	// backup files
	for _, file := range conf.Jobs.Files {
		dest := utils.RemoveTrailingSlash(fmt.Sprintf("%s/%s/%s", conf.Location, date, file.Name))
		fileBackup(file.Paths, file.Ignore, dest, file.Compress)
	}

	// generate checksums
	utils.GenerateChecksums(fmt.Sprintf("%s/%s", conf.Location, date), conf.Checksums)
}

func fileBackup(sources []string, ignore []string, dest string, compress bool) {
	if compress {
		dest = dest + ".tar.gz"
	}

	if _, err := os.Stat(dest); err == nil {
		log.Info(fmt.Sprintf("%s already exists. Skipping", dest))
		return
	}

	if compress {
		// create tar file on dest with contents of sources
		log.Info(fmt.Sprintf("[%s] -> %s\n", strings.Join(sources, ","), dest))
		err := utils.CreateTarball(dest, sources, ignore)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	// copy each source to the destination
	for _, src := range sources {
		// copy src to dest https://opensource.com/article/18/6/copying-files-go
		size, err := cp(src, dest)
		if err != nil {
			log.Debug("%s -> %s (error: %v)\n", src, dest, err)
		} else {
			log.Debug("%s -> %s (size: %d)\n", src, dest, size)
		}
	}
}

func cp(src, dst string) (int64, error) {
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
