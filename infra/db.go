package infra

import (
	"fmt"

	"github.com/my-easy-vault-2026/api-server/lib"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Database modal
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database instance
func NewDatabase(logger lib.Logger, env *lib.Env) Database {
	masterURL := buildDSN(env, env.DBHost, env.DBPort, env.DBUsername, env.DBPassword, env.DBType)

	logger.Info("opening db connection")
	db, err := gorm.Open(mysql.Open(masterURL), &gorm.Config{Logger: logger.GetGormLogger()})
	if err != nil {
		logger.Info("Url: ", masterURL)
		logger.Panic(err)
	}

	logger.Info("creating database if it does't exist")
	if err = db.Exec("CREATE DATABASE IF NOT EXISTS " + env.DBName).Error; err != nil {
		logger.Info("couldn't create database")
		logger.Panic(err)
	}

	logger.Info("using given database")
	if err := db.Exec(fmt.Sprintf("USE %s", env.DBName)).Error; err != nil {
		logger.Info("cannot use the given database")
		logger.Panic(err)
	}
	logger.Info("database connection established")

	// 配置讀寫分離
	if err := setupReadWriteSplitting(db, logger, env); err != nil {
		logger.Info("failed to setup read-write splitting")
		logger.Panic(err)
	}

	// 配置連接池
	sqlDB, err := db.DB()
	if err != nil {
		logger.Panic(err)
	}
	sqlDB.SetMaxOpenConns(env.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(env.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(env.DBConnMaxLifetime)

	logger.Info("database connection established")

	return Database{DB: db}
}

func buildDSN(env *lib.Env, host, port, username, password, dbType string) string {
	// Standard MySQL
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		env.DBName,
	)
}

// setupReadWriteSplitting 配置讀寫分離
func setupReadWriteSplitting(db *gorm.DB, logger lib.Logger, env *lib.Env) error {
	// 如果沒有配置從庫，跳過
	if len(env.DBSlaveHosts) == 0 {
		logger.Info("no slave databases configured, skipping read-write splitting")
		return nil
	}

	// 構建從庫連接
	replicas := make([]gorm.Dialector, 0, len(env.DBSlaveHosts))

	for i, slaveHost := range env.DBSlaveHosts {
		var slavePort string
		var slaveUsername string
		var slavePassword string

		// 從庫可能有自己的端口、用戶名、密碼
		if len(env.DBSlavePorts) > i {
			slavePort = env.DBSlavePorts[i]
		} else {
			slavePort = env.DBPort // 使用主庫端口
		}

		if len(env.DBSlaveUsernames) > i && env.DBSlaveUsernames[i] != "" {
			slaveUsername = env.DBSlaveUsernames[i]
		} else {
			slaveUsername = env.DBUsername // 使用主庫用戶名
		}

		if len(env.DBSlavePasswords) > i && env.DBSlavePasswords[i] != "" {
			slavePassword = env.DBSlavePasswords[i]
		} else {
			slavePassword = env.DBPassword // 使用主庫密碼
		}

		slaveDSN := buildDSN(env, slaveHost, slavePort, slaveUsername, slavePassword, env.DBType)
		logger.Info(fmt.Sprintf("adding slave database %d: %s", i+1, slaveHost))

		replicas = append(replicas, mysql.Open(slaveDSN))
	}

	// 註冊讀寫分離插件
	err := db.Use(dbresolver.Register(dbresolver.Config{
		// 從庫 (讀)
		Replicas: replicas,

		// 負載均衡策略
		Policy: dbresolver.RandomPolicy{}, // 隨機
		// 或使用 dbresolver.RoundRobinPolicy{} // 輪詢

		// 特定表的配置 (可選)
		// Tables: []string{"users", "orders"},
	}).
		// 連接池配置
		SetConnMaxIdleTime(env.DBConnMaxIdleTime).
		SetConnMaxLifetime(env.DBConnMaxLifetime).
		SetMaxIdleConns(env.DBMaxIdleConns).
		SetMaxOpenConns(env.DBMaxOpenConns))

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("read-write splitting configured with %d slave(s)", len(replicas)))
	return nil
}
