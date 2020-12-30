# Go Backup Manager (GBM)
## Features
* file backups
* dockerized OpenLDAP exports
* dockerized MariaDB exports
* ~~dockerized Postgres exports~~
* ~~dockerized MongoDB exports~~
* backup management (e.g. delete backups, but keep the last 2 weeks + ever tenth for the last 6 months)


## Examples
```
$ sudo gbm showConfig
Location: /var/backups/gbm
Checksums: [md5 sha1]
Strategy: Delete backups after 14 days, but ignore [YYYY-MM-10 YYYY-MM-20 YYYY-MM-30] 
Configured File Backups:
- /srv/main -> /var/backups/gbm/2021-01-01/main.tar.gz
- /srv/data -> /var/backups/gbm/2021-01-01/data.tar.gz
- /etc/ssh/sshd_config -> /var/backups/gbm/2021-01-01/sshd_config

Configured LDAP Backups:
- main_ldap_1: ldapsearch -x -D cn=admin,dc=domain,dc=de -W HIDDEN -b dc=domain,dc=de

$ sudo gbm backup
INFO[0000] [/srv/main] -> /var/backups/gbm/2021-01-01/main.tar.gz 
INFO[0148] [/srv/data] -> /var/backups/gbm/2021-01-01/data.tar.gz 
INFO[0439] ldap://main_ldap_1:dc=domain,dc=de -> /var/backups/gbm/2021-01-01/main_ldap_1.ldif 

$ sudo gbm backup
INFO[0000] /var/backups/gbm/2021-01-01/main.tar.gz already exists. Skipping 
INFO[0000] /var/backups/gbm/2021-01-01/data.tar.gz already exists. Skipping 
INFO[0000] ldap://main_ldap_1:dc=domain,dc=de -> /var/backups/gbm/2021-01-01/main_ldap_1.ldif 
```