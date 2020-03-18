package relay

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	Servers  []Server
	Mappings TargetMappings
}

// NewManager returns a new instance of Manager, hydrated with provided mappings
// and servers
func NewManager(mappings TargetMappings) *Manager {
	m := &Manager{
		Mappings: mappings,
	}

	m.loadServers()
	return m
}

// Start iterates over all servers and:
// - Connects each server
// - Sets targets for each server
// - Starts reading from each server
// This will block until all servers have finished reading
func (m *Manager) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, server := range m.Servers {
		serverTargetMappings := m.getServerTargets(server.Name())

		log.Infof("Connecting to server [%s]", server.Name())
		err := server.Connect()
		if err != nil {
			return fmt.Errorf("Failed to connect server %s: %s", server.Name(), err.Error())
		}

		log.Debugf("Settings targets for server [%s]", server.Name())
		err = server.SetTargets(serverTargetMappings)
		if err != nil {
			return fmt.Errorf("Failed to set targets for server %s: %s", server.Name(), err.Error())
		}

		readChannel := make(chan *TargetMessage)
		go m.readChannel(readChannel)

		currentServer := server
		g.Go(func() error {
			log.Debugf("Starting read for server [%s]", currentServer.Name())
			err := currentServer.Read(ctx, readChannel)
			if err != nil {
				return fmt.Errorf("Failed to read from server [%s]: %s", currentServer.Name(), err)
			}

			log.Infof("Finished reading from server [%s]", currentServer.Name())
			return nil
		})
	}

	return g.Wait()
}

func (m *Manager) loadServers() {
	for _, initFunc := range initFuncs {
		m.Servers = append(m.Servers, initFunc()...)
	}
}

func (m *Manager) readChannel(readChannel chan *TargetMessage) {
	for msg := range readChannel {
		m.processMessage(msg)
	}
}

func (m *Manager) processMessage(msg *TargetMessage) {
	serverTargetMappings := m.Mappings.WhereMessageSource(msg)
	for _, serverTargetMapping := range serverTargetMappings {
		for _, server := range m.Servers {
			if serverTargetMapping.To.Server == server.Name() {
				newMessage := NewTargetMessage(msg.Source, serverTargetMapping.To, msg.Payload)
				log.Debugf("Writing message [%s] to target [%s] on server [%s]", newMessage.GetMessage(), serverTargetMapping.To.Name, serverTargetMapping.To.Server)
				err := server.Write(newMessage)
				if err != nil {
					log.Errorf("Failed to write message [%s] to target [%s] on server [%s]: %s", newMessage.GetMessage(), serverTargetMapping.To.Name, serverTargetMapping.To.Server, err.Error())
				}
			}
		}
	}
}

// getServerTargets returns a deduplicated slice of Targets which reference specified server as
// either a source or destination within configured target mappings
func (m *Manager) getServerTargets(server string) []Target {
	var targets []Target

	addTarget := func(target Target) {
		for _, existingTarget := range targets {
			if target.Name == existingTarget.Name {
				return
			}
		}

		targets = append(targets, target)
	}

	for _, serverTargetMapping := range m.Mappings.WhereDestinationServer(server) {
		addTarget(serverTargetMapping.To)
	}

	for _, serverTargetMapping := range m.Mappings.WhereSourceServer(server) {
		addTarget(serverTargetMapping.From)
	}

	return targets
}
