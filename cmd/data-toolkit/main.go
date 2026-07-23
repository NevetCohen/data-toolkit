package main

import (
	"context"
	"fmt"
	"os"

	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
	"data-toolkit/internal/core/datatype"
	fileformat "data-toolkit/internal/fileengine/format"
	"data-toolkit/internal/interfaces/cli"
	logicaloperation "data-toolkit/internal/logical/operation"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	dataTypes, err := datatype.NewBuiltinRegistry()
	if err != nil {
		return err
	}
	defaults, err := config.LoadDefaults()
	if err != nil {
		return err
	}
	service, err := application.New(application.Options{
		DataTypes:  dataTypes,
		Operations: logicaloperation.NewRegistry(logicaloperation.CurrentVersion),
		Formats:    fileformat.NewRegistry(fileformat.CurrentVersion),
		Defaults:   defaults,
	})
	if err != nil {
		return err
	}
	return cli.Execute(ctx, service, args, os.Stdout)
}
