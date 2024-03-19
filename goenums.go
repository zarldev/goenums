package main

import (
	"fmt"
	"os"

	"github.com/zarldev/goenums/pkg/config"
	"github.com/zarldev/goenums/pkg/generator"
)

func main() {
	config, err := ParseInput()
	if err != nil {
		return
	}
	g := generator.New(config)
	err = g.Generate()
	if err != nil {
		fmt.Println("Error generating code", err)
		return
	}
}

func ParseInput() (config.Config, error) {
	err := validateInput()
	if err != nil {
		return config.Config{}, err
	}
	cfgPath := os.Args[1]
	cfg, err := config.ReadConfig(cfgPath)
	if err != nil {
		return config.Config{}, err
	}
	cfg.OutputPath = os.Args[2]
	return cfg, err
}

func validateInput() error {
	if len(os.Args) < 3 {
		printHelp()
		return fmt.Errorf("not enough arguments")
	}
	arg1 := os.Args[1]
	if arg1 == "-h" || arg1 == "--h" || arg1 == "-help" || arg1 == "--help" {
		printHelp()
		return fmt.Errorf("help")
	}
	return nil
}

func printHelp() {
	fmt.Println("Usage: goenums <config file path> <output path>")
}
