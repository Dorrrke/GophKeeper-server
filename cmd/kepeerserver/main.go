package main

import (
	"context"
	"fmt"
	"net"

	"github.com/Dorrrke/GophKeeper-server/internal/config"
	grpcserver "github.com/Dorrrke/GophKeeper-server/internal/grpc"
	"github.com/Dorrrke/GophKeeper-server/internal/logger"
	"github.com/Dorrrke/GophKeeper-server/internal/service"
	"github.com/Dorrrke/GophKeeper-server/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

var (
	// buildVersion - версия сборки.
	buildVersion string
	// buildDate - дата сборки.
	buildDate string
	// buildCommit - комментарии к сборке.
	buildCommit string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	} else {
		fmt.Printf("\n\tBuild version: %s\n", buildVersion)
	}

	if buildDate == "" {
		buildDate = "N/A"
	} else {
		fmt.Printf("\n\tBuild date: %s\n", buildDate)
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	} else {
		fmt.Printf("\n\tBuild commit: %s\n", buildCommit)
	}
	cfg := config.ReadConfig()

	zlog := logger.SetupLogger(cfg.DebugFlag)

	zlog.Debug().Str("db addr", cfg.DBPath).Msg("Creating a database connection")
	conn, err := initDB(cfg.DBPath)
	if err != nil {
		zlog.Panic().Err(err).Msg("Database initialization error")
		panic(err)
	}
	zlog.Debug().Msg("Storage initialization")
	kStor := storage.New(conn, zlog)

	zlog.Debug().Msg("Service initialization")
	kService := service.New(*kStor, zlog)

	zlog.Debug().Msg("gRPC server initialization")
	grpcServer := grpc.NewServer()
	grpcserver.RegisterGrpcServer(grpcServer, kService, zlog)

	zlog.Debug().Str("addr", cfg.ServerAddr).Msg("Create net connection")
	l, err := net.Listen("tcp", cfg.ServerAddr)
	if err != nil {
		zlog.Panic().Err(err).Msg("Init listener error")
		panic(err)
	}

	zlog.Debug().Msg("GophKeeper server started")
	err = grpcServer.Serve(l)
	if err != nil {
		zlog.Panic().Err(err).Msg("Serve error")
		panic(err)
	}
}

func initDB(DBAddr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), DBAddr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
