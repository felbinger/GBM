package cmd

import (
	"fmt"
	"gbm/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
	"strconv"
	"time"
)

// showConfigCmd represents the showConfig command
var showConfigCmd = &cobra.Command{
	Use:   "showConfig",
	Short: "Show configuration",
	Run: func(cmd *cobra.Command, args []string) {
		showCreds, err := strconv.ParseBool((cmd.Flag("show-creds").Value).String())
		if err != nil {
			log.Fatal("Unable to parse secure parameter. Exiting...")
		}
		showConfiguration(Configuration, showCreds)
	},
}

func init() {
	showConfigCmd.PersistentFlags().BoolP("show-creds", "s", false, "Show credentials in configuration")
	rootCmd.AddCommand(showConfigCmd)
}


// showConfiguration show the passed config The secure parameter can be used to hide passwords.
func showConfiguration(conf utils.Config, showCreds bool) {
	var hiddenPasswords bool

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

	fmt.Println("\nConfigured Database Backups:")
	for _, dbms := range []string{"MariaDB", "Postgres", "MongoDB"} {
		for _, db := range conf.GetJobs(dbms) {
			if db.Container.IsEmpty() {
				fmt.Printf("- %s: %s can't be reached!\n", dbms, db.ContainerName)
				continue
			}
			var credentials string
			if db.Auth {
				if showCreds {
					credentials = fmt.Sprintf("%s:%s@", db.Username, db.Password)
				} else {
					hiddenPasswords = true
					credentials = fmt.Sprintf("%s:***@", db.Username)
				}
			}
			fmt.Printf("- %s: %s%s/%v\n", dbms, credentials, db.ContainerName, db.Databases)
		}
	}

	fmt.Println("\nConfigured LDAP Backups:")
	for _, ldap := range conf.Jobs.Ldap {
		if ldap.Container.IsEmpty() {
			fmt.Printf("- %s can't be reached!\n", ldap.ContainerName)
			continue
		}
		var password string
		if showCreds {
			password = ldap.BindPw
		} else {
			hiddenPasswords = true
			password = "***"
		}
		fmt.Printf("- %s: ldapsearch -x -D %s -W %s -b %s\n", ldap.ContainerName, ldap.BindDn, password, ldap.BaseDn)

		// show disclaimer
		if hiddenPasswords {
			fmt.Println("You can use the option [-s|--show-creds] to see the configured passwords.")
		}
	}
}
