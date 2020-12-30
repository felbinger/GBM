package main

import (
	"gbm/cmd"
	"gbm/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

func main() {
	if strings.TrimSpace(utils.GetProcessOwner()) != "root" {
		log.Fatal("gbm requires root access! Exiting...")
	}
	cmd.Execute()
}
