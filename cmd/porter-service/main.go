package main

import (
	//"context"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"time"

	//"os/signal"
	//"runtime/debug"
	"path/filepath"
	"strings"

	//"get.porter.sh/porter/pkg/cli"
	//"get.porter.sh/porter/pkg/config"
	//"get.porter.sh/porter/pkg/porter/version"
	"get.porter.sh/porter/pkg/grpc"
	//"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/signals"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	//"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	//"go.opentelemetry.io/otel/attribute"
	go_grpc "google.golang.org/grpc"
)

var includeDocsCommand = false

//go:embed helptext/usage.txt
var usageText string

const (
	cmdName string = "porter-service"
	// Indicates that config should not be loaded for this command.
	// This is used for commands like help and version which should never
	// fail, even if porter is misconfigured.
	skipConfig string = "skipConfig"
)

func main() {
	// flags definition
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	fs.String("host", "", "Host to bind service to")
	fs.Int("port", 9898, "HTTP port to bind service to")
	fs.Int("secure-port", 0, "HTTPS port")
	fs.Int("port-metrics", 0, "metrics port")
	fs.Int("grpc-port", 0, "gRPC port")
	fs.String("grpc-service-name", "porter", "gPRC service name")
	fs.String("level", "info", "log level debug, info, warn, error, fatal or panic")
	//fs.StringSlice("backend-url", []string{}, "backend service URL")
	//fs.Duration("http-client-timeout", 2*time.Minute, "client timeout duration")
	//fs.Duration("http-server-timeout", 30*time.Second, "server read and write timeout duration")
	//fs.Duration("server-shutdown-timeout", 5*time.Second, "server graceful shutdown timeout duration")
	fs.String("data-path", "/data", "data local path")
	fs.String("config-path", "", "config dir path")
	fs.String("cert-path", "/data/cert", "certificate path for HTTPS port")
	fs.String("config", "config.yaml", "config file name")
	//fs.String("ui-path", "./ui", "UI local path")
	//fs.String("ui-logo", "", "UI logo")
	//fs.String("ui-color", "#34577c", "UI color")
	//fs.String("ui-message", fmt.Sprintf("greetings from podinfo v%v", version.VERSION), "UI message")
	//fs.Bool("h2c", false, "allow upgrading to H2C")
	//fs.Bool("random-delay", false, "between 0 and 5 seconds random delay by default")
	//fs.String("random-delay-unit", "s", "either s(seconds) or ms(milliseconds")
	//fs.Int("random-delay-min", 0, "min for random delay: 0 by default")
	//fs.Int("random-delay-max", 5, "max for random delay: 5 by default")
	//fs.Bool("random-error", false, "1/3 chances of a random response error")
	//fs.Bool("unhealthy", false, "when set, healthy state is never reached")
	//fs.Bool("unready", false, "when set, ready state is never reached")
	//fs.Int("stress-cpu", 0, "number of CPU cores with 100 load")
	//fs.Int("stress-memory", 0, "MB of data to load into memory")
	//fs.String("cache-server", "", "Redis address in the format 'tcp://<host>:<port>'")
	fs.String("otel-service-name", "", "service name for reporting to open telemetry address, when not set tracing is disabled")
	versionFlag := fs.BoolP("version", "v", false, "get version number")

	// parse flags
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		fs.PrintDefaults()
		os.Exit(2)
	case *versionFlag:
		fmt.Println("version")
		os.Exit(0)
	}

	// bind flags and environment variables
	viper.BindPFlags(fs)
	hostname, _ := os.Hostname()
	viper.Set("hostname", hostname)
	//viper.Set("version", version.VERSION)
	viper.Set("version", "version")
	viper.Set("revision", "revision")
	viper.SetEnvPrefix("PORTER_SVC")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// load config from file
	if _, fileErr := os.Stat(filepath.Join(viper.GetString("config-path"), viper.GetString("config"))); fileErr == nil {
		viper.SetConfigName(strings.Split(viper.GetString("config"), ".")[0])
		viper.AddConfigPath(viper.GetString("config-path"))
		if readErr := viper.ReadInConfig(); readErr != nil {
			fmt.Printf("Error reading config file, %v\n", readErr)
		}
	}

	// configure logging
	logger, _ := initZap(viper.GetString("level"))
	defer logger.Sync()
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	// load gRPC server config
	var grpcCfg grpc.Config
	if err := viper.Unmarshal(&grpcCfg); err != nil {
		logger.Panic("config unmarshal failed", zap.Error(err))
	}

	// start gRPC server
	var grpcServer *go_grpc.Server
	if grpcCfg.Port > 0 {
		grpcSrv, _ := grpc.NewServer(&grpcCfg, logger)
		grpcServer = grpcSrv.ListenAndServe()
	}
	// log version and port
	logger.Info("Starting porter-server",
		zap.String("version", viper.GetString("version")),
		zap.String("revision", viper.GetString("revision")),
		zap.String("port", strconv.Itoa(grpcCfg.Port)),
	)

	// graceful shutdown
	stopCh := signals.SetupSignalHandler()
	serverShutdownTimeout := time.Duration(time.Second * 30)
	sd, _ := signals.NewShutdown(serverShutdownTimeout, logger)
	sd.Graceful(stopCh, grpcServer)
}

func initZap(logLevel string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	}

	zapEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
