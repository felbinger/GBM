package cmd

import (
	"context"
	"fmt"
	"gbm/utils"
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

var regenerateChecksums bool

func backup(conf utils.Config) {
	date := time.Now().Format("2006-01-02")
	err := os.MkdirAll(fmt.Sprintf("%s/%s", conf.Location, date), 0755)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create backup directory: %s", conf.Location))
		return
	}

	// backup files
	for _, file := range conf.Jobs.Files {
		dest := utils.RemoveTrailingSlash(fmt.Sprintf("%s/%s/%s", conf.Location, date, file.Name))
		fileBackup(file.Paths, file.Ignore, file.IgnoreExtension, dest, file.Compress)
	}

	// backup ldap
	for _, ldap := range conf.Jobs.Ldap {
		dest := fmt.Sprintf("%s/%s/%s.ldif", conf.Location, date, ldap.ContainerName)
		err := ldapBackup(dest, ldap)
		if err != nil {
			log.Println(err)
		}
	}

	// backup databases
	for _, dbms := range []string{"MariaDB", "Postgres", "MongoDB"} {
		for _, db := range conf.GetJobs(dbms) {
			if db.Container.IsEmpty() {
				log.Debug(fmt.Sprintf("- %s: %s can't be reached!\n", dbms, db.ContainerName))
				continue
			}

			dest := fmt.Sprintf("%s/%s/%s/", conf.Location, date, db.ContainerName)
			err := databaseBackup(dest, db, dbms)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// generate checksums
	if regenerateChecksums {
		utils.GenerateChecksums(fmt.Sprintf("%s/%s", conf.Location, date), conf.Checksums)
	}
}

func fileBackup(sources []string, ignore []string, ignoreExt []string, dest string, compress bool) {
	if compress {
		dest = dest + ".tar.gz"
	}

	if _, err := os.Stat(dest); err == nil {
		log.Info(fmt.Sprintf("%s already exists. Skipping", dest))
		return
	}
	regenerateChecksums = true

	if compress {
		// create tar file on dest with contents of sources
		log.Info(fmt.Sprintf("[%s] -> %s\n", strings.Join(sources, ","), dest))
		err := utils.CreateTarball(dest, sources, ignore, ignoreExt)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	// copy each source to the destination
	for _, src := range sources {
		// copy src to dest https://opensource.com/article/18/6/copying-files-go
		size, err := utils.Cp(src, dest)
		if err != nil {
			log.Info(fmt.Sprintf("%s -> %s (error: %v)\n", src, dest, err))
		} else {
			log.Info(fmt.Sprintf("%s -> %s (size: %d)\n", src, dest, size))
		}
	}
}

func ldapBackup(dest string, ldap utils.Ldap) error {
	ctx := context.Background()

	if _, err := os.Stat(dest); err == nil {
		log.Info(fmt.Sprintf("%s already exists. Skipping", dest))
		return nil
	}

	if ldap.Container.IsEmpty() {
		log.Info(fmt.Sprintf("%s cannot be reached. Skipping", ldap.ContainerName))
		return nil
	}
	regenerateChecksums = true

	log.Info(fmt.Sprintf("ldap://%s:%s -> %s\n", ldap.ContainerName, ldap.BaseDn, dest))

	cmd := []string{"ldapsearch", "-x", "-D", ldap.BindDn, "-w", ldap.BindPw, "-b", ldap.BaseDn}
	resp, err := utils.Exec(ctx, ldap.ContainerName, cmd)
	if err != nil {
		return err
	}

	// write stdout to file
	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.Write(resp)
	if err != nil {
		return err
	}

	return nil
}

func databaseBackup(dest string, db utils.Database, dbms string) error {
	err := os.MkdirAll(dest, 0775)
	if err != nil {
		return err
	}

	switch dbms {
	case "MariaDB":
		for _, database := range db.Databases {
			dbDest := fmt.Sprintf("%s%s.sql", dest, database)

			if _, err := os.Stat(dbDest); err == nil {
				log.Info(fmt.Sprintf("%s already exists. Skipping", dbDest))
				return nil
			}

			if db.Container.IsEmpty() {
				log.Info(fmt.Sprintf("%s cannot be reached. Skipping", db.ContainerName))
				return nil
			}

			regenerateChecksums = true

			log.Info(fmt.Sprintf("%s/%s -> %s\n", db.ContainerName, database, dbDest))

			cmd := []string{
				"mysqldump",
				"--lock-tables",
				"--protocol", "tcp",
				"--host", "localhost",
				"--port", "3306",
			}
			if db.Auth {
				cmd = append(cmd, []string {"--user", db.Username, "--password=" + db.Password}...)
			}
			cmd = append(cmd, database)
			resp, err := utils.Exec(context.Background(), db.ContainerName, cmd)
			if err != nil {
				return err
			}

			// write stdout to file
			dst, err := os.Create(dbDest)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = dst.Write(resp)
			if err != nil {
				return err
			}

		}
	case "Postgres":
	case "MongoDB":
		for _, database := range db.Databases {
			dbDest := fmt.Sprintf("%s%s.tar", dest, database)

			if _, err := os.Stat(dbDest); err == nil {
				log.Info(fmt.Sprintf("%s already exists. Skipping", dbDest))
				return nil
			}
			regenerateChecksums = true

			log.Info(fmt.Sprintf("%s/%s -> %s\n", db.ContainerName, database, dbDest))

			cmd := []string{
				"mongodump",
				"--host", "localhost",
				"--port", "27017",
				"--db=" + database,
			}
			if db.Auth {
				cmd = append(cmd, []string {
					"--username", db.Username,
					"--password=" + db.Password,
					"--authenticationDatabase", "admin",
					"--authenticationMechanism", "SCRAM-SHA-1",
				}...)
			}
			cmd = append(cmd, database)
			resp, err := utils.Exec(context.Background(), db.ContainerName, cmd)
			if err != nil {
				return err
			}

			// write stdout to file
			dst, err := os.Create(dbDest)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = dst.Write(resp)
			if err != nil {
				return err
			}

		}
	}
	return nil
}