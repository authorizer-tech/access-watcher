package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	watchpb "github.com/authorizer-tech/access-watcher/gen/go/authorizer-tech/accesswatcher/v1alpha1"
	watcher "github.com/authorizer-tech/access-watcher/internal"
	"github.com/authorizer-tech/access-watcher/internal/datastores/postgres"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var grpcPort = flag.Int("grpc-port", 50052, "The bind port for the grpc server")
var httpPort = flag.Int("http-port", 8082, "The bind port for the grpc-gateway http server")
var configPath = flag.String("config", "./localconfig/config.yaml", "The path to the server config")

type Config struct {
	GrpcGateway struct {
		Enabled bool
	}

	Postgres struct {
		Host     string
		Port     int
		Database string
	}
}

func main() {

	flag.Parse()

	viper.SetConfigFile(*configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to load server config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to Unmarshal server config: %v", err)
	}

	pgUsername := viper.GetString("POSTGRES_USERNAME")
	pgPassword := viper.GetString("POSTGRES_PASSWORD")
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		pgUsername,
		pgPassword,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
	)

	pool, err := pgxpool.Connect(context.TODO(), dsn)
	if err != nil {
		log.Fatalf("Failed to establish a connection to Postgres database: %v", err)
	}

	store, err := postgres.NewChangelogDatastore(pool)
	if err != nil {
		log.Fatalf("Failed to initialize the Changelog datastore: %v", err)
	}

	log.Info("Starting access-watcher")
	log.Infof("  Version: %s", version)
	log.Infof("  Date: %s", date)
	log.Infof("  Commit: %s", commit)
	log.Infof("  Go version: %s", runtime.Version())

	watcherOpts := []watcher.AccessWatcherOption{
		watcher.WithChangelogDatastore(store),
	}
	w, err := watcher.NewAccessWatcher(watcherOpts...)
	if err != nil {
		log.Fatalf("Failed to initialize the access-watcher: %v", err)
	}

	addr := fmt.Sprintf(":%d", *grpcPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start the TCP listener on '%v': %v", addr, err)
	}

	grpcOpts := []grpc.ServerOption{}
	server := grpc.NewServer(grpcOpts...)
	watchpb.RegisterWatchServiceServer(server, w)

	go func() {
		reflection.Register(server)

		log.Infof("Starting grpc server at '%v'..", addr)

		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to start the gRPC server: %v", err)
		}
	}()

	var gateway *http.Server
	if cfg.GrpcGateway.Enabled {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Register gRPC server endpoint
		// Note: Make sure the gRPC server is running properly and accessible
		mux := gwruntime.NewServeMux()

		opts := []grpc.DialOption{grpc.WithInsecure()}

		if err := watchpb.RegisterWatchServiceHandlerFromEndpoint(ctx, mux, addr, opts); err != nil {
			log.Fatalf("Failed to initialize grpc-gateway WatchService handler: %v", err)
		}

		gateway = &http.Server{
			Addr:    fmt.Sprintf(":%d", *httpPort),
			Handler: mux,
		}

		go func() {
			log.Infof("Starting grpc-gateway server at '%v'..", gateway.Addr)

			// Start HTTP server (and proxy calls to gRPC server endpoint)
			if err := gateway.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("Failed to start grpc-gateway HTTP server: %v", err)
			}
		}()
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	<-exit

	log.Info("Shutting Down..")

	if cfg.GrpcGateway.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := gateway.Shutdown(ctx); err != nil {
			log.Errorf("Failed to gracefully shutdown the grpc-gateway server: %v", err)
		}
	}

	server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := w.Close(ctx); err != nil {
		log.Errorf("Failed to gracefully close the access-watcher: %v", err)
	}

	pool.Close()

	log.Info("Shutdown Complete. Goodbye 👋")
}
