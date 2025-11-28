package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type config struct {
	Port        int    `mapstructure:"PORT"`
	ClerkSecret string `mapstructure:"CLERK_SECRET_KEY"`
	Terminal    string `mapstructure:"TERMINAL"`
}

func NewConfigFromEnv() (config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error reading .env file: %v", err)
		}
	}

	viper.SetConfigFile(".env.local")
	viper.SetConfigType("env")
	if err := viper.MergeInConfig(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error reading .env.local file: %v", err)
		}
	}

	viper.AutomaticEnv()
	var cfg config
	if err := viper.BindEnv("PORT"); err != nil {
		return cfg, fmt.Errorf("failed to bind env PORT: %w", err)
	}
	if err := viper.BindEnv("CLERK_SECRET_KEY"); err != nil {
		return cfg, fmt.Errorf("failed to bind env CLERK_SECRET_KEY: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.ClerkSecret == "" {
		return cfg, fmt.Errorf("CLERK_SECRET_KEY is required")
	}

	return cfg, nil
}
