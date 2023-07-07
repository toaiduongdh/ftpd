package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

func main() {
	// Parse command line arguments
	broker := flag.String("broker", "", "Kafka broker address")
	topic := flag.String("topic", "", "Kafka topic")
	inputFile := flag.String("inputfile", "", "Input file containing messages to publish")
	useTls := flag.Bool("tls", false, "Input file containing messages to publish")
	caRootFile := flag.String("ca-root", "", "ca root file if needed")
	username := flag.String("username", "", "")
	password := flag.String("password", "", "")
	useSasl := flag.Bool("sasl", false, "")
	decodeBase64 := flag.Bool("base64decode", false, "")
	batch := flag.Int("batch", 100, "number of msg per batch producing")

	flag.Parse()

	// Validate command line arguments
	if *broker == "" || *topic == "" || *inputFile == "" {
		fmt.Println("broker, topic, and inputfile are required parameters")
		os.Exit(1)
	}

	// Create Kafka producer
	config := sarama.NewConfig()
	config.Net.TLS.Enable = *useTls
	config.Producer.Return.Successes = true
	if *caRootFile != "" {
		raw, err := os.ReadFile(*caRootFile)
		if err != nil {
			panic(err)
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(raw)
		if !ok {
			panic("failed appending certs from pem somehow")
		}
		tlsConfig := &tls.Config{RootCAs: certPool}
		config.Net.TLS.Config = tlsConfig
	}
	if *useSasl {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = *username
		config.Net.SASL.Password = *password
		config.Net.SASL.Handshake = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	}
	producer, err := sarama.NewSyncProducer([]string{*broker}, config)
	if err != nil {
		fmt.Printf("Failed to create Kafka producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// Read messages from input file and publish to Kafka topic
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Failed to open input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	msges := []*sarama.ProducerMessage{}
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)

		payload := parts[1]
		if *decodeBase64 {
			raw, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				panic(err)
			}
			payload = string(raw)
		}
		message := &sarama.ProducerMessage{
			Topic: *topic,
			Key:   sarama.StringEncoder(parts[0]),
			Value: sarama.ByteEncoder(payload),
		}
		msges = append(msges, message)
		if len(msges) == *batch {
			err = producer.SendMessages(msges)
			if err != nil {
				fmt.Printf("Failed to publish message: %v\n", err)
				os.Exit(1)
			}
			msges = msges[:0]
		}
	}
	if len(msges) > 0 {
		err = producer.SendMessages(msges)
		if err != nil {
			fmt.Printf("Failed to publish message: %v\n", err)
			os.Exit(1)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read input file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All messages published successfully")
}
