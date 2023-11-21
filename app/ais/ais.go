package ais

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
)

// Slice to store the message parts in
var messageParts = make(map[int]map[int]string)

func HandleMessage(message string) error {
	log.Printf("Received a new message: %s\n", message)
	prefix, numberOfMessageParts, partNumber, messageID, err := simpleParse(message)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	if prefix != "!AIVDM" {
		return fmt.Errorf("invalid prefix: %s", prefix)
	}

	if numberOfMessageParts == 1 {
		log.Printf("Message is complete: %s\n", message)

		packet, err := decodeCompleteMessage(message)
		if err != nil {
			return fmt.Errorf("failed to handle message: %v", err)
		}

		log.Printf("Decoded message successfully, packet: %s\n", packet.MessageType)

		// TODO: This
		// err = events.SendMessage(message)
		// if err != nil {
		// return fmt.Errorf("failed to handle message: %v", err)
		// }
	} else {
		log.Printf("Message is not complete yet: %s\n", message)
		err := addMessagePart(messageID, partNumber, message)
		if err != nil {
			return fmt.Errorf("failed to handle message part: %v", err)
		}

		if isMessageComplete(messageID, numberOfMessageParts) {
			allMessages := getMultipartMessages(messageID)
			removeCompleteMessage(messageID)
			if allMessages == nil {
				return fmt.Errorf("failed to retrieve complete message")
			}

			packet, err := decodeCompleteMessages(allMessages)
			if err != nil {
				return fmt.Errorf("failed to handle message: %v", err)
			}

			log.Printf("Decoded multipart message successfully, packet: %s\n", packet.MessageType)

			// TODO: This:
			// err = events.SendMessage(message)
			// if err != nil {
			// 	return fmt.Errorf("failed to handle message: %v", err)
			// }
		}
	}

	return nil
}

func simpleParse(message string) (string, int, int, int, error) {
	parts := strings.Split(message, ",")
	if len(parts) < 4 {
		return "", 0, 0, 0, fmt.Errorf("invalid message format")
	}

	prefix, rawNumberOfMessageParts, rawPartNumber, rawMessageID := parts[0], parts[1], parts[2], parts[3]

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
			fmt.Printf("%+v\n\n", decoded)
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

func getMultipartMessages(messageID int) []string {
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
