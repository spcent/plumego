package tasks

import "encoding/json"

// SendTaskPayload is the queue payload for SMS sends.
type SendTaskPayload struct {
	MessageID string `json:"message_id"`
	TenantID  string `json:"tenant_id"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Provider  string `json:"provider,omitempty"`
}

const SendTopic = "sms.send"

func EncodeSendTask(payload SendTaskPayload) ([]byte, error) {
	return json.Marshal(payload)
}

func DecodeSendTask(data []byte) (SendTaskPayload, error) {
	var payload SendTaskPayload
	if len(data) == 0 {
		return payload, nil
	}
	err := json.Unmarshal(data, &payload)
	return payload, err
}
