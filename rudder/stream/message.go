package stream

import "encoding/json"

const (
	pulsarKeyMessageID       = "messageID"
	pulsarKeyRoutingKey      = "routingKey"
	pulsarKeyWorkspaceID     = "workspaceID"
	pulsarKeySourceID        = "sourceID"
	pulsarKeyUserID          = "userID"
	pulsarKeySourceJobRunID  = "sourceJobRunID"
	pulsarKeySourceTaskRunID = "sourceTaskRunID"
	pulsarKeyTraceID         = "traceID"
)

type Message struct {
	Properties MessageProperties `json:"properties"`
	Payload    json.RawMessage   `json:"payload"`
}

type MessageProperties struct {
	MessageID       string `json:"messageID"`
	RoutingKey      string `json:"routingKey"`
	WorkspaceID     string `json:"workspaceID"`
	UserID          string `json:"userID"`
	SourceID        string `json:"sourceID"`
	SourceJobRunID  string `json:"sourceJobRunID"`  //optional
	SourceTaskRunID string `json:"sourceTaskRunID"` //optional
	TraceID         string `json:"traceID"`
}

// FromPulsarMessage converts a Pulsar message to a Message.
func FromPulsarMessage(properties map[string]string, payload []byte) (Message, error) {
	return Message{
		Properties: MessageProperties{
			MessageID:       properties[pulsarKeyMessageID],
			RoutingKey:      properties[pulsarKeyRoutingKey],
			WorkspaceID:     properties[pulsarKeyWorkspaceID],
			UserID:          properties[pulsarKeyUserID],
			SourceID:        properties[pulsarKeySourceID],
			SourceJobRunID:  properties[pulsarKeySourceJobRunID],
			SourceTaskRunID: properties[pulsarKeySourceTaskRunID],
			TraceID:         properties[pulsarKeyTraceID],
		},
		Payload: json.RawMessage(payload),
	}, nil
}

// ToPulsarMessage converts a Message to a Pulsar message.
func ToPulsarMessage(msg Message) (map[string]string, []byte) {
	properties := msg.Properties
	return map[string]string{
		pulsarKeyMessageID:       properties.MessageID,
		pulsarKeyRoutingKey:      properties.RoutingKey,
		pulsarKeyWorkspaceID:     properties.WorkspaceID,
		pulsarKeySourceID:        properties.SourceID,
		pulsarKeyUserID:          properties.UserID,
		pulsarKeySourceJobRunID:  properties.SourceJobRunID,
		pulsarKeySourceTaskRunID: properties.SourceTaskRunID,
		pulsarKeyTraceID:         properties.TraceID,
	}, []byte(msg.Payload)
}
