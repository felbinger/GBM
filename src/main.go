package main

import (
	"github.com/felbinger/GBM/cmd"
	"github.com/felbinger/GBM/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

func main() {
	if strings.TrimSpace(utils.GetProcessOwner()) != "root" {
		log.Fatal("GBM requires root access! Exiting...")
	}
	cmd.Execute()
}
