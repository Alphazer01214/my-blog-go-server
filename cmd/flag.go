package cmd

import "flag"

func InitFlag() {
	postgresMigrate := flag.Bool("migrate", true, "migrate database")
	flag.Parse()

	if *postgresMigrate {
		if err := MigrateDB(); err != nil {
			panic(err)
		}
	}
}
