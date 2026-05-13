package cmd

import "flag"

func InitFlag() {
	flag.Parse()

	postgresMigrate := flag.Bool("migrate", true, "migrate database")

	if *postgresMigrate {
		if err := MigrateDB(); err != nil {
			panic(err)
		}
	}
}
