package ais

import (
	"ais-receiver/events"
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
	prefix, numberOfMessageParts, partNumber, messageID, err := simpleParse(message)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
	}

	if prefix != "!AIVDM" {
		return fmt.Errorf("invalid prefix: %s", prefix)
	}

	allowMessageTypes := []uint8{1, 2, 3, 5, 18, 24, 27}

	if numberOfMessageParts == 1 {
		packet, err := decodeCompleteMessage(message)
		if err != nil {
			//log.Println(packet.MessageType) // TODO: REMOVE THIS TEMPORARY LINE
			return fmt.Errorf("failed to handle message: %v", err)
		}

		for _, allowedMessageType := range allowMessageTypes {
			if packet.Packet.GetHeader().MessageID == allowedMessageType {
				err = events.SendMessage(message)
				if err != nil {
					return fmt.Errorf("failed to handle message: %v", err)
				}

				log.Println("Ok", packet.Packet.GetHeader())
				break
			}
		}

	} else {
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
				log.Println(packet.MessageType) // TODO: REMOVE THIS TEMPORARY LINE
				return fmt.Errorf("failed to handle message: %v", err)
			}

			// TODO: Send useful data to the event hub
			err = events.SendMessage(message)
			if err != nil {
				return fmt.Errorf("failed to handle message: %v", err)
			}

			log.Println("Ok", packet.Packet.GetHeader())
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
