package ais

import (
	"ais-receiver/ais/Policy"
	"ais-receiver/events"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
)

// Slice to store the message parts in
var messageParts = make(map[int]map[int]string)

func MessageReceiver(message string) error {
	prefix, numberOfMessageParts, partNumber, messageID, err := ParseMessage(message)

	if err != nil {
		return err
	}

	if err := Policy.AssertMessageIsAivdm(prefix); err != nil {
		return err
	}

	return handleMessage(message, numberOfMessageParts, messageID, partNumber)
}

func ParseMessage(message string) (string, int, int, int, error) {
	prefix, numberOfMessageParts, partNumber, messageID, err := simpleParse(message)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to parse message: %v", err)
	}
	return prefix, numberOfMessageParts, partNumber, messageID, nil
}

func handleMessage(message string, numberOfMessageParts int, messageID int, partNumber int) error {
	if numberOfMessageParts == 1 {
		return handleSinglePartMessage(message)
	} else {
		return collectMessageParts(message, numberOfMessageParts, messageID, partNumber)
	}
}

func collectMessageParts(message string, numberOfMessageParts int, messageID int, partNumber int) error {
	if err := handleMessagePart(message, messageID, partNumber); err != nil {
		return err
	}

	if isMessageComplete(messageID, numberOfMessageParts) {
		return handleCompleteMultiPartMessage(messageID)
	}
	return nil
}

func handleCompleteMultiPartMessage(messageID int) error {
	fullMessage := getMultipartMessage(messageID)
	removeCompleteMessage(messageID)
	if fullMessage == nil {
		return fmt.Errorf("failed to retrieve complete message")
	}

	rawPacket, err := decodeCompleteMessages(fullMessage)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	return handleRawPacket(rawPacket)
}

func handleRawPacket(rawPacket *aisnmea.VdmPacket) error {
	if isAllowedMessagePacket(rawPacket) {
		err := handlePacket(rawPacket)
		if err != nil {
			return fmt.Errorf("failed to handle message: %v", err)
		}
	}
	return nil
}

func handleMessagePart(message string, messageID int, partNumber int) error {
	err := addMessagePart(messageID, partNumber, message)
	if err != nil {
		return fmt.Errorf("failed to handle message part: %v", err)
	}
	return nil
}

func handleSinglePartMessage(message string) error {
	rawPacket, err := decodeCompleteMessage(message)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	if isAllowedMessagePacket(rawPacket) {
		err = handlePacket(rawPacket)
		if err != nil {
			return fmt.Errorf("failed to handle message: %v", err)
		}
	}
	return nil
}

func handlePacket(rawPacket *aisnmea.VdmPacket) error {
	packet := rawPacket.Packet
	if packet == nil {
		return fmt.Errorf("raw packet is empty: %v", rawPacket)
	}

	return packetToJson(packet)
}

func packetToJson(packet ais.Packet) error {
	jsonPacket, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	if err := events.SendMessage(string(jsonPacket)); err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	return nil
}

func isAllowedMessagePacket(packet *aisnmea.VdmPacket) bool {
	allowedMessageTypes := []uint8{1, 2, 3, 5, 18, 24, 27}
	return slices.Contains(allowedMessageTypes, packet.Packet.GetHeader().MessageID)
}

func simpleParse(message string) (string, int, int, int, error) {
	parts := strings.Split(message, ",")
	if len(parts) < 7 {
		return "", 0, 0, 0, fmt.Errorf("invalid message format: %s", message)
	}

	prefix, rawNumberOfMessageParts, rawPartNumber, rawMessageID, _, payload := parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
	if payload == "" {
		return "", 0, 0, 0, fmt.Errorf("payload from message is empty: %s", message)
	}

	numberOfMessageParts, err := strconv.Atoi(rawNumberOfMessageParts)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to parse numberOfMessageParts: %w", err)
	}

	partNumber, err := strconv.Atoi(rawPartNumber)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to parse partNumber: %w", err)
	}

	messageID, err := strconv.Atoi(rawMessageID)
	if err != nil && numberOfMessageParts > 1 {
		return "", 0, 0, 0, fmt.Errorf("failed to parse messageID: %w", err)
	}

	return prefix, numberOfMessageParts, partNumber, messageID, nil
}

func decodeCompleteMessages(messages []string) (*aisnmea.VdmPacket, error) {
	nm := aisnmea.NMEACodecNew(ais.CodecNew(false, false))

	for _, message := range messages {
		decoded, err := nm.ParseSentence(message)
		if err != nil {
			return nil, fmt.Errorf("failed to decode message: %w", err)
		}

		if decoded != nil {
			return decoded, nil
		}
	}

	return nil, fmt.Errorf("failed to decode message :(")
}

func decodeCompleteMessage(message string) (*aisnmea.VdmPacket, error) {
	return decodeCompleteMessages(
		[]string{message},
	)
}

func addMessagePart(messageID, partNumber int, message string) error {
	if _, ok := messageParts[messageID]; !ok {
		messageParts[messageID] = make(map[int]string)
	}
	messageParts[messageID][partNumber] = message

	return nil
}

func isMessageComplete(messageID, numberOfMessageParts int) bool {
	if _, ok := messageParts[messageID]; !ok {
		return false
	}

	return len(messageParts[messageID]) == numberOfMessageParts
}

func getMultipartMessage(messageID int) []string {
	if _, ok := messageParts[messageID]; !ok {
		return nil
	}

	parts := messageParts[messageID]
	var completeMessage []string
	for i := 1; i <= len(parts); i++ {
		completeMessage = append(completeMessage, parts[i])
	}

	return completeMessage
}

func removeCompleteMessage(messageID int) {
	delete(messageParts, messageID)
}
