package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type File struct {
	Name     string
	Compress bool
	Paths    []string
	Ignore   []string
}

type Job struct {
	Files []File
}

type Strategy struct {
	ExpiryDays int `yaml:"expiry_days"`
	Ignore     []string
}

type Config struct {
	Location  string
	Checksums []string
	Jobs      Job
	Strategy  Strategy
}

func Configure(fileName string) *Config {

	c := Config{}

	// check if fileName exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("%s does not exist!\n", fileName))
		os.Exit(1)
	}

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(fmt.Sprintf("Configuration Error:   #%v ", err))
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		log.Fatal(fmt.Sprintf("Configuration Error: %v", err))
	}

	// remove trailing slash in location path
	c.Location = RemoveTrailingSlash(c.Location)

	return &c
}

func RemoveTrailingSlash(path string) string {
	if path[len(path)-1:] == "/" {
		path = path[:len(path)-1]
	}
	return path
}

// ShowConfiguration show the passed config The secure parameter can be used to hide passwords.
func ShowConfiguration(conf Config, secure bool) {

	fmt.Println("Location: " + conf.Location)
	fmt.Printf("Checksums: %v\n", conf.Checksums)
	fmt.Printf("Strategy: Delete backups after %d days, but ignore %v ",
		conf.Strategy.ExpiryDays, conf.Strategy.Ignore)

	fmt.Println("\nConfigured File Backups:")
	for _, file := range conf.Jobs.Files {
		for _, src := range file.Paths {
			date := time.Now().Format("2006-01-02")
			dest := path.Join(conf.Location, date)

			destName := file.Name
			if file.Compress {
				destName += ".tar.gz"
			}
			dest = path.Join(dest, destName)

			fmt.Printf("- %s -> %s\n", src, dest)
		}
		for _, src := range file.Ignore {
			log.Debug(fmt.Sprintf("- %s will be ignored.\n", src))
		}
	}
}
