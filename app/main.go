package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
)

// Slice to store the message parts in
var messageParts = make(map[int]map[int]string)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Println(err)
	}

	defer listener.Close()

	for {
		acceptConnection(listener)
	}
}

func acceptConnection(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Client connected %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		err := handleMessage(message)
		if err != nil {
			log.Println(err)
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Println("Error reading:", err)
	}
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

func handleMessage(message string) error {
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

		// TODO: sendMessageToEventHub()
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
			// TODO: sendMessageToEventHub()
		}
	}

	if 1 == 2 {
		err := sendMessageToEventHub(message)
		if err != nil {
			return fmt.Errorf("failed to handle message: %v", err)
		}
	}

	return nil
}

func sendMessageToEventHub(aisMessage string) error {
	producerClient, err := createProducerClient()
	if err != nil {
		return fmt.Errorf("failed to send message to event hub: %w", err)
	}

	defer producerClient.Close(context.TODO())

	err = sendMessageAsBatch(producerClient, aisMessage)
	if err != nil {
		return fmt.Errorf("failed to send message to event hub: %w", err)
	}

	return nil
}

func createProducerClient() (*azeventhubs.ProducerClient, error) {
	connectionString := os.Getenv("ENDPOINT_CONNECTION_STRING")
	eventHubName := os.Getenv("EVENT_HUB_NAME")

	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(connectionString, eventHubName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer client: %w", err)
	}

	return producerClient, nil
}

func sendMessageAsBatch(producerClient *azeventhubs.ProducerClient, aisMessage string) error {
	batch, err := createEventBatch(producerClient)
	if err != nil {
		return fmt.Errorf("failed to send message as batch: %w", err)
	}

	err = fillEventBatch(batch, aisMessage)
	if err != nil {
		return fmt.Errorf("failed to send message as batch: %w", err)
	}

	err = sendBatchToEventHub(batch, producerClient)
	if err != nil {
		return fmt.Errorf("failed to send message as batch: %w", err)
	}

	return nil
}

func sendBatchToEventHub(batch *azeventhubs.EventDataBatch, producerClient *azeventhubs.ProducerClient) error {
	err := producerClient.SendEventDataBatch(context.Background(), batch, nil)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

func fillEventBatch(batch *azeventhubs.EventDataBatch, aisMessage string) error {
	events := messageToEvents(aisMessage)
	for _, event := range events {
		err := batch.AddEventData(event, nil)
		if err != nil {
			return fmt.Errorf("failed to add event data to batch: %w", err)
		}
	}

	return nil
}

func createEventBatch(producerClient *azeventhubs.ProducerClient) (*azeventhubs.EventDataBatch, error) {
	newBatchOptions := &azeventhubs.EventDataBatchOptions{}
	batch, err := producerClient.NewEventDataBatch(context.TODO(), newBatchOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create event batch: %w", err)
	}

	return batch, nil
}

func messageToEvents(aisMessage string) []*azeventhubs.EventData {
	return []*azeventhubs.EventData{
		{
			Body: []byte(aisMessage),
		},
	}
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
