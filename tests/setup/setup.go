package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	viper.SetDefault("config", "../../configs/dev/chorus.yaml")

	pflag.String("config", viper.GetString("config"), "The configuration file path can be relative or absolute.")
	pflag.Bool("clean", false, "Clean database after tests.")
	err := viper.BindPFlag("config", pflag.Lookup("config"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("clean", pflag.Lookup("clean"))
	if err != nil {
		panic(err)
	}
	pflag.Parse()

	viper.SetConfigFile(viper.GetString("config"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("config file not found:", viper.GetString("config"))
		os.Exit(1)
	} else {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}

	if _, err := logger.InitLoggers(provider.ProvideConfig()); err != nil {
		fmt.Println("unable to initialize loggers:", err.Error())
		os.Exit(1)
	}

	dbID := "chorus"
	cfg := provider.ProvideConfig().Storage.Datastores[dbID]
	ctx := context.Background()

	// Failsafe: if the dabase does not start with the 'acc_' prefix, return an error so we don't wipe accidently the INT databases.
	if !strings.HasPrefix(cfg.Database, "acc_") {
		logger.TechLog.Fatal(ctx, "DB must start with the 'acc_' prefix", zap.String("db_name", cfg.Database))
	}

	db := provider.ProvideDB(dbID)
	if db == nil {
		logger.TechLog.Fatal(ctx, "cannot connect to DB", zap.String("db_name", cfg.Database))
	}

	switch cfg.Type {
	case provider.POSTGRES:
		if viper.GetBool("clean") {
			logger.TechLog.Info(ctx, "clean database", zap.String("db_name", cfg.Database))
			deletePostgresDB(ctx, db.DB.GetSqlxDB(), cfg.Database)
			logger.TechLog.Info(ctx, "database cleaned", zap.String("db_name", cfg.Database))
		} else {
			logger.TechLog.Info(ctx, "setup database", zap.String("db_name", cfg.Database))
			deletePostgresDB(ctx, db.DB.GetSqlxDB(), cfg.Database)
			createPostgresDB(ctx, db.DB.GetSqlxDB(), cfg.Database)
			logger.TechLog.Info(ctx, "setup done", zap.String("db_name", cfg.Database))
		}
	default:
		logger.TechLog.Fatal(ctx, "unsupported storage", zap.String("storage_type", cfg.Type))
	}
	logger.TechLog.Info(ctx, "setup complete")
}

func deletePostgresDB(ctx context.Context, db *sqlx.DB, dbName string) {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE;", dbName))
	if err != nil {
		log.Fatalf("could not delete DB %s: %s", dbName, err.Error())
	}
}
func createPostgresDB(ctx context.Context, db *sqlx.DB, dbName string) {
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if err != nil {
		log.Fatalf("could not create DB %s: %s", dbName, err.Error())
	}
}
