package main

import (
	"api-server/bootstrap"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = bootstrap.RootApp.Execute()
}
