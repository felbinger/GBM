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

// deleteCmd represents the deleteBackup command
var deleteCmd = &cobra.Command{
	Use:   "deleteBackup",
	Short: "Manage Mode",
	Run: func(cmd *cobra.Command, args []string) {
		noInput, err := strconv.ParseBool(cmd.Flag("no-input").Value.String())
		if err != nil {
			log.Fatal("Unable to parse no-input parameter. Exiting...")
		}
		deleteBackup(Configuration, noInput)
	},
}

func init() {
	deleteCmd.PersistentFlags().Bool("no-input", false, "Don't ask for confirmation (required for cronjob)")
	rootCmd.AddCommand(deleteCmd)
}

func deleteBackup(conf utils.Config, noInput bool) {
	// get a list of existing backups
	files, err := ioutil.ReadDir(conf.Location)
	if err != nil {
		log.Fatal(err)
	}

	// generate a list of backups to deleteBackup (older then n days and not in the list to be ignored)
	var deletableBackups []string
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
			splitFileName := strings.Split(fileName, "-")
			splitPattern := strings.Split(pattern, "-")

			if splitFileName[0] == splitPattern[0] ||
				splitFileName[1] == splitPattern[1] ||
				splitFileName[2] == splitPattern[2] {
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
		deletableBackups = append(deletableBackups, fileName)
	}

	if len(deletableBackups) == 0 {
		log.Debug("Nothing to deleteBackup...")
		return
	}

	if !noInput {
		// ask user if deletion of backups is what he would like to do
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Type 'YES' to confirm the deletion of the following backups:\n%s\nYour selection: ",
			strings.Join(deletableBackups, ", "))
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if strings.Split(confirmation, "\n")[0] != "YES" {
			return
		}
	}
	// deleteBackup backups
	for _, date := range deletableBackups {
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
