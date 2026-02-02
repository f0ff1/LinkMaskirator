package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"LinkMaskirator/service"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &cli.App{
		Name:     "Link Maskirator",
		Usage:    "Программа для маскировки ссылок внутри текста",
		Version:  "0.7.0",
		Compiled: time.Now(),

		Authors: []*cli.Author{
			{
				Name: "Андрей Кантур",
			},
		},

		Flags: []cli.Flag{
			//log-level
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Value:   "Info",
				Usage:   "Уровень логирования (debug|info|warn|error)",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Value:   false,
				Usage:   "Вывод логов в формате JSON",
			},
		},

		Action: func(c *cli.Context) error {
			cli.ShowAppHelp(c)
			return nil
		},

		Commands: []*cli.Command{
			{
				Name:    "mask",
				Aliases: []string{"m"},
				Usage:   "Маскировка ссылок в тексте",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "source",
						Aliases:  []string{"s"},
						Usage:    "Путь к файлу с ссылками",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "dest",
						Aliases:  []string{"d"},
						Usage:    "Путь к конечному файлу",
						Value:    "C:/GoLand/GoCourse/LinkMaskirator/txtFiles/maskedLinks.txt",
						Required: false,
					},
					&cli.IntFlag{
						Name:    "workers",
						Aliases: []string{"wc"},
						Value:   10,
						Usage:   "Количество горутин для обработки",
					},
					&cli.BoolFlag{
						Name:    "slowmode",
						Aliases: []string{"sm"},
						Value:   false,
						Usage:   "Замедлить работу маскиратора для последующей проверки GS",
					},
					&cli.IntFlag{
						Name:    "timeout",
						Aliases: []string{"t"},
						Value:   5,
						Usage:   "Таймаут выполнения программы. Чаще используется в паре с --slowmode",
					},
				},

				Action: maskAction,
			},
		},

		Metadata: map[string]interface{}{
			"app_ctx": ctx,
		},

		//Действие перед выполением любой команды
		Before: func(c *cli.Context) error {
			//Настраиваем логгер перед выполнением команды

			setupLogger(c.String("log-level"), c.Bool("json"))
			return nil
		},
		//Действие после выполнения команды (даже если была ошибка)
		After: func(c *cli.Context) error {
			ellapsed := time.Since(c.App.Compiled)
			slog.Info("команда выполена",
				"command", c.Command.Name,
				"duration (ms)", ellapsed.Milliseconds())
			return nil
		},
	}

	//Запускаем приложение (с graceful shutdown)
	var appErr error
	appDone := make(chan struct{}, 1)

	go func() {
		defer close(appDone)
		appErr = app.Run(os.Args)
	}()

	select {
	case <-appDone:
		slog.Info("приложение завершилось самостоятельно (Без GS)")
	case signal := <-signalChan:
		slog.Info("получил сигнал graceful shutdown",
			"signal", signal.String(),
			"time", time.Now())
		cancel()
		slog.Debug("контекст приложения отменен")
		select {
		case <-appDone:
			slog.Info("graceful shutdown выполнен успешно")
		case <-time.After(10 * time.Second):
			slog.Warn("таймаут graceful shutdown (10 секунд). Завершил принудительно.")
		}
	}

	if appErr != nil {
		slog.Error("приложение завершено с ошибкой",
			"error", appErr)
		os.Exit(1)
	}

	slog.Info("приложение успешно завершено")

}

func setupLogger(levelStr string, jsonOutput bool) {
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	//Выбор формата вывода (человеческий или json)
	var handler slog.Handler
	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	slog.SetDefault(slog.New(handler))

	slog.Debug("Логгер настроен",
		"level", levelStr,
		"json output", jsonOutput)

}

func runMaskingProcess(ctx context.Context, inputFile, outputFile string, workers int, slowmode bool) error {
	if workers < 0 {
		return fmt.Errorf("количество воркеров должно быть положительным: %d", workers)
	}

	if inputFile == "" {
		return fmt.Errorf("не указан исходный файл")
	}

	slog.DebugContext(ctx, "создание сервиса", "workers", workers, "max goroutines", runtime.NumCPU())
	factory := service.NewServiceFactory(workers, slowmode)

	svc := factory.CreateMaskService(inputFile, outputFile)

	return svc.Run(ctx)
}

func maskAction(c *cli.Context) error {
	inputFile := c.String("source")
	outputFile := c.String("dest")
	countWorkers := c.Int("workers")
	isSlowMode := c.Bool("slowmode")
	timeOut := c.Int("timeout")

	if timeOut < 1 {
		return fmt.Errorf("Ошибка длительности таймаута. Таймаут не может быть меньше 1 секунды")
	}

	appCtx, ok := c.App.Metadata["app_ctx"].(context.Context)
	if !ok {
		appCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(appCtx, time.Duration(timeOut)*time.Second)
	defer cancel()

	slog.InfoContext(ctx, "начало маскировки",
		"input", inputFile,
		"output", outputFile,
		"count workers", countWorkers,
		"slow mode status", isSlowMode,
		"timeout", timeOut)

	err := runMaskingProcess(
		ctx,
		inputFile,
		outputFile,
		countWorkers,
		isSlowMode,
	)

	if err != nil {
		slog.ErrorContext(ctx, "ошибка при маскировке",
			"error", err)

		if ctx.Err() == context.DeadlineExceeded {
			slog.InfoContext(ctx, "Время таймаута истекло", "timeout (s)", timeOut)
			return cli.Exit("Превышено время ожидания", 2)
		}
		return cli.Exit(fmt.Sprintf("Ошибка маскировки: %v", err), 1)
	}

	slog.InfoContext(ctx, "маскировка завершена успешно", "output file", c.String("dest"))
	return nil

}
