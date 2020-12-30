package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Ldap struct {
	Container     Container `yaml:"-"`
	ContainerName string    `yaml:"container_name"`
	BaseDn        string    `yaml:"base_dn"`
	BindDn        string    `yaml:"bind_dn"`
	BindPw        string    `yaml:"bind_pw"`
}

type Database struct {
	Type          string    `yaml:"-"`
	Container     Container	`yaml:"-"`
	ContainerName string    `yaml:"container_name"`
	Auth 	      bool      `yaml:"-"`
	Username      string	`yaml:",omitempty"`
	Password      string	`yaml:",omitempty"`
	Databases     []string
}

type Config struct {
	Location  string
	Checksums []string
	Jobs      struct {
		Files []struct {
			Name            string
			Compress        bool
			Paths           []string
			Ignore          []string
			IgnoreExtension []string `yaml:"ignore_extension"`
		}
		Ldap     []Ldap
		MariaDb  []Database
		Postgres []Database
		MongoDb  []Database
	}
	Strategy struct {
		ExpiryDays int `yaml:"expiry_days"`
		Ignore     []string
	}
}

func (c Config) GetJobs(dbms string) []Database {
	switch dbms {
	case "MariaDB":
		return c.Jobs.MariaDb
	case "Postgres":
		return c.Jobs.Postgres
	case "MongoDB":
		return c.Jobs.MongoDb
	}
	return []Database{}
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

	// OpenLDAP
	for i, ldap := range c.Jobs.Ldap {
		c.Jobs.Ldap[i].Container = GetContainerByName(ldap.ContainerName)
	}

	// Databases
	for i, db := range c.Jobs.MariaDb {
		c.Jobs.MariaDb[i].Container = GetContainerByName(db.ContainerName)
		fmt.Println(db.Username)
		if db.Username != "" && db.Password != "" {
			c.Jobs.MariaDb[i].Auth = true
		}
	}

	return &c
}

func RemoveTrailingSlash(path string) string {
	if path[len(path)-1:] == "/" {
		path = path[:len(path)-1]
	}
	return path
}

func AppendTrailingSlash(path string) string {
	if path[len(path)-1:] != "/" {
		path += "/"
	}
	return path
}
