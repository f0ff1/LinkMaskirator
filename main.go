package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli/v2"

	"LinkMaskirator/service"
)

func main() {
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
				Usage:   "Маскировка ссылок в тексте. [-s: путь к файлу с ссылками. -d: путь к конечному файлу. -wc: количество воркеров. -sm: slowmode]",
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
						Usage:    "Путь к конечному файлу (если не указан, по умолчанию - txtFiles/maskedLinks.txt)",
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
				},

				Action: maskAction,
			},
		},
		//Дейцствие перед выполением любой команды
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

	// Запуск приложения
	if err := app.Run(os.Args); err != nil {
		slog.Error("Ошибка выполнения", "error", err)
		os.Exit(1) //Выход с кодом ошибки
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.InfoContext(ctx, "начало маскировки",
		"input", c.String("source"),
		"output", c.String("dest"))

	inputFile := c.String("source")
	outputFile := c.String("dest")
	countWorkers := c.Int("workers")
	isSlowMode := c.Bool("slowmode")

	slog.DebugContext(ctx, "DEBUG: получен флаг slowmode из CLI",
		"slowmode", isSlowMode,
		"raw_flag", c.String("slowmode"),
		"is_set", c.IsSet("slowmode"))

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
			return cli.Exit("Превышено время ожидания (30 секунд)", 2)
		}
		return cli.Exit(fmt.Sprintf("Ошибка маскировки: %v", err), 1)
	}

	slog.InfoContext(ctx, "маскировка завершена успешно", "output file", c.String("dest"))
	return nil

}
