package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
	"log"
	"net"
)

// TODO: Replace with environment variables (and add these environment variables to the container instance)
const (
	connectionString = "Endpoint=sb://nspeventhubs.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=5IlHGz3aULWsHg3KH4tjN4jatLjL9kc6G+AEhJdt/Jk="
	eventHubName     = "ais-event-hub"
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
		go handleMessage(message)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading:", err)
	}
}

type aisData struct {
	Nmae uint32
	Type uint8
}

func handleMessage(message string) {
	log.Printf("Received AIS message: %s\n", message)

	if message[0] != '!' && message[0] != '$' {
		return
	}

	nm := aisnmea.NMEACodecNew(ais.CodecNew(false, false))

	log.Printf("NMEACodecNew: %v\n", nm)

	decoded, err := nm.ParseSentence(message)
	if err != nil {
		log.Fatalf("failed to decode NMEA sentence: %s", err)
	}

	log.Printf("Decoded: %v\n", decoded)

	//aisData := aisData{
	//	Nmae: decoded.Packet.GetHeader().UserID,
	//	Type: decoded.Packet.GetHeader().MessageID,
	//}
	//
	//log.Printf("AIS data: %v\n", aisData.Nmae)

	err = sendMessageToEventHub(message)
	if err != nil {
		log.Fatalf("failed to send message to event hub: %s", err)
	}
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
