package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
)

func GetProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(stdout)
}
