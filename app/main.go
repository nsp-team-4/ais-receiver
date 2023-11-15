package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Println("Error reading:", err)
	}
}

func handleMessage(message string) error {
	log.Printf("Received AIS message: %s\n", message)

	err := sendMessageToEventHub(message)
	if err != nil {
		return fmt.Errorf("failed to handle message: %v", err)
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

	log.Println("Batch sent successfully")

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
