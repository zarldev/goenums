package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zarldev/goenums/pkg/config"
	"github.com/zarldev/goenums/pkg/generator"
)

func main() {
	cfg, err := parseInput()
	if err != nil {
		logrus.Fatalf("error parse input: %v", err)
	}

	g := generator.New(cfg)
	err = g.Generate()
	if err != nil {
		logrus.Fatalf("error generating code: %v", err)
	}
}

func parseInput() (config.Config, error) {
	var pkgArg string
	var typeArg string
	var valuesArg string
	var outputArg string
	var cfgArg string

	flag.StringVar(&pkgArg, "package", "", "Package enum that will be generated. E.g -package \"validation\"")
	flag.StringVar(&typeArg, "type", "", "Enum type. E.g -type \"status\"")
	flag.StringVar(&valuesArg, "values", "", "Enum values seperated by \",\". E.g -values \"Failed, Passed, Skipped, Scheduled, Running\"")
	flag.StringVar(&outputArg, "output", "", "Output path that will be generated. E.g -output \"./output\"")
	flag.StringVar(&cfgArg, "cfg", "", "Config file path. E.g -cfg \"./input.json\"")

	flag.Parse()

	var enumsArg []string
	if valuesArg != "" {
		for _, v := range strings.Split(valuesArg, ",") {
			enumsArg = append(enumsArg, strings.ReplaceAll(v, " ", ""))
		}
	}

	var cfg config.Config
	if cfgArg == "" {
		cfg.Configs = []config.EnumConfig{
			{Package: pkgArg, Type: typeArg, Enums: enumsArg},
		}
	} else {
		var err error
		cfg, err = config.ReadConfig(cfgArg)
		if err != nil {
			return config.Config{}, fmt.Errorf("error read config file: %w", err)
		}
	}

	cfg.OutputPath = outputArg

	err := cfg.Validate()
	if err != nil {
		return config.Config{}, fmt.Errorf("error validate config: %w", err)
	}

	return cfg, nil
}
