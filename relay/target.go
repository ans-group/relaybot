package relay

import "fmt"

// TargetMessage represents a message between Source and Destination
type TargetMessage struct {
	Source      Target
	Destination Target
	Payload     TargetMessagePayload
}

// NewTargetMessage returns a pointer to an intialised TargetMessage
func NewTargetMessage(source Target, destination Target, payload TargetMessagePayload) *TargetMessage {
	return &TargetMessage{
		Source:      source,
		Destination: destination,
		Payload:     payload,
	}
}

// GetMessage returns a formatted string representation of TargetMessage
func (m *TargetMessage) GetMessage() string {
	return fmt.Sprintf("(%s@%s) %s", m.Payload.User, m.Source.Server, m.Payload.Msg)
}

// TargetMessagePayload represents the payload of a TargetMessage
type TargetMessagePayload struct {
	User string
	Msg  string
}

// TargetMappings represents a slice of TargetMapping structs
type TargetMappings []TargetMapping

// WhereDestinationServer returns a slice of TargetMappings which reference specified server as
// a destination within configured target mappings
func (t *TargetMappings) WhereDestinationServer(server string) []TargetMapping {
	var targetMappings []TargetMapping
	for _, targetMapping := range *t {
		if targetMapping.To.Server == server {
			targetMappings = append(targetMappings, targetMapping)
		}
	}

	return targetMappings
}

// WhereSourceServer returns a slice of TargetMappings which reference specified server as
// a source within configured target mappings
func (t *TargetMappings) WhereSourceServer(server string) []TargetMapping {
	var targetMappings []TargetMapping
	for _, targetMapping := range *t {
		if targetMapping.From.Server == server {
			targetMappings = append(targetMappings, targetMapping)
		}
	}

	return targetMappings
}

// WhereMessageSource returns a slice of TargetMappings which reference source in specified message msg
func (t *TargetMappings) WhereMessageSource(msg *TargetMessage) []TargetMapping {
	var targetMappings []TargetMapping
	for _, targetMapping := range *t {
		if targetMapping.From.Server == msg.Source.Server && targetMapping.From.Name == msg.Source.Name {
			targetMappings = append(targetMappings, targetMapping)
		}
	}

	return targetMappings
}

// TargetMapping represents a one-to-one mapping of targets
type TargetMapping struct {
	To   Target
	From Target
}

// Target represents a single target
type Target struct {
	Server string
	Name   string
}
