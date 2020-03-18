package main

import (
	"context"

	"github.com/ukfast/relaybot/relay"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	log.Info("Reading configuration")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	// Setup logging
	if viper.GetBool("debug") {
		log.Info("Setting logging TraceLevel")
		log.SetLevel(log.TraceLevel)

		log.Info("Setting relay debug")
		relay.Debug = true
	}

	// Load mappings from configuration file
	log.Info("Loading mappings")
	var targetMappings relay.TargetMappings
	err = viper.UnmarshalKey("mappings", &targetMappings)
	if err != nil {
		log.Fatalf("Failed to load server target mappings: %s", err.Error())
	}

	// Create manager with retrieved mappings
	m := relay.NewManager(targetMappings)

	// Start manager (blocking)
	err = m.Start(context.Background())
	if err != nil {
		log.Fatalf("Manager failed: %s", err.Error())
	}
}
