package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"ride-hail/internal/shared/config"
	"ride-hail/pkg/logger"
)

func Run(ctx context.Context, config config.Config) error {
	args := os.Args[1:]

	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	log, err := logger.InitLogger(config.LogLevel, true)

	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	slog.SetDefault(log)

	command := args[0]
	switch command {
	case "driver":
		return DriverRun(ctx, config)
	case "rider":
		return RideRun(ctx, config)
	case "admin":
		return AdminRun(ctx, config)
	default:
		printHelp()
		return fmt.Errorf("unknown command: %s", command)
	}

}
func printHelp() {
	helpText := `Usage: ride-hail [command]

Commands:
  driver	Start the application in driver mode
  rider		Start the application in rider mode
  admin	 	Start the application in admin mode

Use "ride-hail [command] --help" for more information about a command.
`
	fmt.Println(helpText)
}
