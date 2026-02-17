package main

import (
	"github.com/joho/godotenv"
	"github.com/my-easy-vault-2026/api-server/bootstrap"
)

// @title Easy Vault API
// @version 1.0
// @description Easy Vault API Server
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /test
// @schemes http https
func main() {
	_ = godotenv.Load()
	_ = bootstrap.RootApp.Execute()
}
