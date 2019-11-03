package application

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	flag "github.com/spf13/pflag"
	"integration_framework/application_config"
	"integration_framework/helper"
	"integration_framework/plugins"
	"integration_framework/testing"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var shutdownRequested bool

func New() (*Application, error) {
	args := new(LaunchArgs)
	flag.BoolVar(&args.runOnce, "once", false, "run all tests once and exit")
	flag.Parse()
	args.configurationPath = flag.Arg(0)

	err := args.Validate()
	if err != nil {
		return nil, fmt.Errorf("unable to validate args: %v", err)
	}

	tmpDirectory, err := filepath.Abs("./integration_environment")
	if err != nil {
		return nil, fmt.Errorf("unable to get absolute path to tmp directory: %v", err)
	}

	application := Application{
		args:         args,
		tmpDirectory: tmpDirectory,
	}
	application.setupGracefulExit()
	return &application, nil
}

type Application struct {
	tmpDirectory      string
	args              *LaunchArgs
	config            *application_config.Config
	parsedConfig      *testing.ParsedConfig
	launcherType      string
	launcher          plugins.ILauncher
	shutdownRequested bool
	configUpdated     bool
}

type LaunchArgs struct {
	configurationPath string
	runOnce           bool
}

func (a *Application) Start() error {
	err := helper.EnsureDirectory(a.tmpDirectory)
	if err != nil {
		return fmt.Errorf("unable to ensure directory: %v", err)
	}

	if !a.args.runOnce {
		err = a.startWatcher()
		if err != nil {
			return fmt.Errorf("unable to start config watcher")
		}
	}

	exitCode, err := a.start()
	log.Println(">> returned from start", exitCode, err)

	time.Sleep(2 * time.Second)
	a.shutdown()

	os.Exit(exitCode)
	return nil
}

func (a *Application) startWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("unable to create watcher: %v", err)
	}

	err = application_config.IterateOverConfigFiles(a.args.configurationPath, func(filename string) error {
		err := watcher.Add(filename)
		if err != nil {
			return fmt.Errorf("unable to add file %q to watcher: %v", filename, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to add path to watcher: %v", err)
	}

	go func() {
		defer watcher.Close()
		for {
			if a.shutdownRequested {
				return
			}

			select {
			// watch for events
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("--> config update detected!")
					a.configUpdated = true
					return
				}

				// watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("watcher error", err)
			}
		}
	}()

	return nil
}

func (a *Application) start() (exitCode int, err error) {
	for {
		fmt.Println("--> start The Cycle")

		cwd, err := os.Getwd()
		if err != nil {
			return 1, fmt.Errorf("unable to get cwd: %v", err)
		}
		config, err := application_config.LoadConfig(a.args.configurationPath, cwd)
		if err != nil {
			return 1, fmt.Errorf("unable to load config: %v", err)
		}
		a.config = config

		parsedConfig, err := testing.ParseConfig(config)
		if err != nil {
			return 1, fmt.Errorf("unable to parse config: %v", err)
		}
		a.parsedConfig = parsedConfig

		if a.launcher != nil && a.launcherType != config.Launcher {
			err := a.launcher.Shutdown()
			if err != nil {
				return 1, fmt.Errorf("unable to shutdown launcher: %v", err)
			}
			a.launcher = nil
		}
		if a.launcher == nil {
			launcher, err := plugins.NewLauncher(config.Launcher, a.tmpDirectory)
			if err != nil {
				return 1, fmt.Errorf("unable to create launcher: %v", err)
			}
			a.launcher = launcher
			a.launcherType = config.Launcher
		}

		err = a.launcher.ConfigUpdated(a.config, parsedConfig.Services, parsedConfig.EnvironmentInitializers)
		if err != nil {
			return 1, fmt.Errorf("unable to re-launch tests: %v", err)
		}

		for serviceName, service := range parsedConfig.Services {
			err := service.Start()
			if err != nil {
				return 1, fmt.Errorf("unable to start service %q: %v", serviceName, err)
			}
		}

		exitCode = parsedConfig.RunTests()

		if a.args.runOnce {
			return exitCode, nil
		}

		for {
			time.Sleep(5 * time.Second)
			if a.configUpdated {
				a.configUpdated = false
				err := a.startWatcher()
				if err != nil {
					return 1, fmt.Errorf("unable to re-create watcher: %v", err)
				}
				break
			}
			if a.shutdownRequested {
				return 0, nil
			}
		}
	}
}

func (a *Application) setupGracefulExit() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		<-interruptChan
		log.Println("\n\n----> Received an interrupt")
		a.shutdown()
	}()
}

func (a *Application) shutdown() {
	a.shutdownRequested = true
	if a.launcher != nil {
		err := a.launcher.Shutdown()
		if err != nil {
			log.Printf("unable to stop launcher: %v", err)
		}
		a.launcher = nil
	}
}

func (a *LaunchArgs) Validate() error {
	if a.configurationPath == "" {
		return fmt.Errorf("path to configuration should be specified")
	}
	return nil
}
