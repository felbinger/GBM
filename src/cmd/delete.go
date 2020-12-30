package cmd

import (
	"bufio"
	"fmt"
	"github.com/felbinger/GBM/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Manage Mode",
	Run: func(cmd *cobra.Command, args []string) {
		noInput, err := strconv.ParseBool(cmd.Flag("no-input").Value.String())
		if err != nil {
			log.Fatal("Unable to parse no-input parameter. Exiting...")
		}
		delete(Configuration, noInput)
	},
}

func init() {
	deleteCmd.PersistentFlags().Bool("no-input", false, "Don't ask for confirmation (required for cronjob)")
	rootCmd.AddCommand(deleteCmd)
}

func delete(conf utils.Config, noInput bool) {
	// get a list of existing backups
	files, err := ioutil.ReadDir(conf.Location)
	if err != nil {
		log.Fatal(err)
	}

	// generate a list of backups to delete (older then n days and not in the list to be ignored)
	var deleteableBackups []string
	date := time.Now().AddDate(0, 0, -(conf.Strategy.ExpiryDays))
	for _, file := range files {
		fileName := file.Name()

		// skip backups of the latest n days
		if fileName > date.Format("2006-01-02") {
			//fmt.Printf("Keep (less than %d days old): %s\n", conf.Strategy.KeepFor, fileName)
			continue
		}

		// check for ignore pattern
		found := false
		for _, pattern := range conf.Strategy.Ignore {
			splittedFileName := strings.Split(fileName, "-")
			splittedPattern := strings.Split(pattern, "-")

			if splittedFileName[0] == splittedPattern[0] ||
				splittedFileName[1] == splittedPattern[1] ||
				splittedFileName[2] == splittedPattern[2] {
				//fmt.Printf("Keep (matches ignore pattern): %s\n", fileName)
				found = true
				break
			}
		}

		// reset indicator variable
		if found {
			found = false
			continue
		}
		deleteableBackups = append(deleteableBackups, fileName)
	}

	if len(deleteableBackups) == 0 {
		log.Debug("Nothing to delete...")
		return
	}

	if !noInput {
		// ask user if deletion of backups is what he would like to do
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Type 'YES' to confirm the deletion of the following backups:\n%s\nYour selection: ",
			strings.Join(deleteableBackups, ", "))
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if strings.Split(confirmation, "\n")[0] != "YES" {
			return
		}
	}
	// delete backups
	for _, date := range deleteableBackups {
		err := remove(fmt.Sprintf("%s/%s", conf.Location, date))
		if err != nil {
			fmt.Println(err)
		}
	}
	log.Info("Backups have been deleted successfully. Exiting...")
}

func remove(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	err = os.Remove(dir)
	if err != nil {
		return err
	}
	return nil
}
