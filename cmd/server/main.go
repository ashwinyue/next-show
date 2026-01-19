// Package main 提供服务器入口.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mervyn/next-show/internal/biz"
	handler "github.com/mervyn/next-show/internal/handler/http"
	"github.com/mervyn/next-show/internal/model"
	"github.com/mervyn/next-show/internal/store"
)

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	// 自动迁移（开发环境）
	if viper.GetBool("database.auto_migrate") {
		if err := autoMigrate(db); err != nil {
			log.Fatalf("failed to auto migrate: %v", err)
		}
	}

	// 依赖注入
	s := store.NewStore(db)
	b := biz.NewBiz(s)
	h := handler.NewHandler(b)

	// 初始化 Gin
	if viper.GetString("server.mode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 注册路由
	h.RegisterRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf(":%d", viper.GetInt("server.port"))
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.auto_migrate", false)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		log.Println("config file not found, using defaults")
	}

	// 环境变量覆盖
	viper.AutomaticEnv()
	return nil
}

func initDB() (*gorm.DB, error) {
	dsn := viper.GetString("database.dsn")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			viper.GetString("database.host"),
			viper.GetInt("database.port"),
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.dbname"),
			viper.GetString("database.sslmode"),
		)
	}

	logLevel := logger.Silent
	if viper.GetString("server.mode") == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(viper.GetInt("database.max_idle_conns"))
	sqlDB.SetMaxOpenConns(viper.GetInt("database.max_open_conns"))
	sqlDB.SetConnMaxLifetime(time.Duration(viper.GetInt("database.conn_max_lifetime")) * time.Second)

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Provider{},
		&model.Agent{},
		&model.AgentRelation{},
		&model.Session{},
		&model.Message{},
		&model.Checkpoint{},
		&model.CheckpointEvent{},
		&model.MCPServer{},
		&model.MCPTool{},
		&model.AgentTool{},
	)
}
