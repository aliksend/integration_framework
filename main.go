package main

import (
	"integration_framework/application"
	_ "integration_framework/plugins/docker_compose"
	_ "integration_framework/plugins/filesystem"
	_ "integration_framework/plugins/graphql"
	_ "integration_framework/plugins/http_server"
	_ "integration_framework/plugins/mysql"
	_ "integration_framework/plugins/postgres"
	_ "integration_framework/plugins/smtp"
	"log"
	"math/rand"
	"time"
)

// TODO хорошо бы вынести нейминг сервисов и их файлов из самих сервисов в docker_compose плагин чтобы он диктовал условия, а не плагины

func main() {
	rand.Seed(time.Now().UnixNano())

	app, err := application.New()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Start()
	if err != nil {
		log.Fatal(err)
	}
}
