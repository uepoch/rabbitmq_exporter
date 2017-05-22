package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/imdario/mergo"

	"strings"
)

var (
	config rabbitExporterConfig
)

var (
	fRabbitURL      = flag.String("rabbit.url", "", "")
	fRabbitUsername = flag.String("rabbit.user", "", "")
	fRabbitPassword = flag.String("rabbit.password", "", "")
	fPublishPort    = flag.String("web.listen-address", "127.0.0.1:9090", "")
	fSkipQueues     = flag.String("collector.skip", "", "")
	fIncludeQueues  = flag.String("collector.include", "", "")
)

type rabbitExporterConfig struct {
	RabbitURL          string
	RabbitUsername     string
	RabbitPassword     string
	PublishPort        string
	OutputFormat       string
	CAFile             string
	InsecureSkipVerify bool
	SkipQueues         string
	IncludeQueues      string
	RabbitCapabilities rabbitCapabilitySet
}

type rabbitCapability string
type rabbitCapabilitySet map[rabbitCapability]bool

const (
	rabbitCapNoSort rabbitCapability = "no_sort"
	rabbitCapBert   rabbitCapability = "bert"
)

var allRabbitCapabilities = rabbitCapabilitySet{
	rabbitCapNoSort: true,
	rabbitCapBert:   true,
}

func initConfig() {
	defaultConfig := rabbitExporterConfig{
		RabbitURL:          "http://localhost:15672",
		RabbitUsername:     "guest",
		RabbitPassword:     "guest",
		PublishPort:        "9090",
		OutputFormat:       "TTY", //JSON
		CAFile:             "ca.pem",
		InsecureSkipVerify: false,
		SkipQueues:         "^$",
		IncludeQueues:      ".*",
		RabbitCapabilities: make(rabbitCapabilitySet),
	}
	if url := os.Getenv("RABBIT_URL"); url != "" {
		if valid, _ := regexp.MatchString("https?://[a-zA-Z.0-9]+", strings.ToLower(url)); valid {
			defaultConfig.RabbitURL = url
		}
	}

	if user := os.Getenv("RABBIT_USER"); user != "" {
		defaultConfig.RabbitUsername = user
	}

	if pass := os.Getenv("RABBIT_PASSWORD"); pass != "" {
		defaultConfig.RabbitPassword = pass
	}

	if port := os.Getenv("PUBLISH_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			defaultConfig.PublishPort = port
		}

	}
	if output := os.Getenv("OUTPUT_FORMAT"); output != "" {
		defaultConfig.OutputFormat = output
	}

	if cafile := os.Getenv("CAFILE"); cafile != "" {
		defaultConfig.CAFile = cafile
	}
	if insecureSkipVerify := os.Getenv("SKIPVERIFY"); insecureSkipVerify == "true" || insecureSkipVerify == "1" {
		defaultConfig.InsecureSkipVerify = true
	}

	if SkipQueues := os.Getenv("SKIP_QUEUES"); SkipQueues != "" {
		defaultConfig.SkipQueues = SkipQueues
	}

	if IncludeQueues := os.Getenv("INCLUDE_QUEUES"); IncludeQueues != "" {
		defaultConfig.IncludeQueues = IncludeQueues
	}

	if rawCapabilities := os.Getenv("RABBIT_CAPABILITIES"); rawCapabilities != "" {
		defaultConfig.RabbitCapabilities = parseCapabilities(rawCapabilities)
	}
	config = rabbitExporterConfig{
		RabbitURL:      *fRabbitURL,
		RabbitUsername: *fRabbitUsername,
		RabbitPassword: *fRabbitPassword,
		PublishPort:    *fPublishPort,
	}
	if err := mergo.Merge(&config, defaultConfig); err != nil {
		log.Fatalf("%v while merging configs", err)
	}
}

func parseCapabilities(raw string) rabbitCapabilitySet {
	result := make(rabbitCapabilitySet)
	candidates := strings.Split(raw, ",")
	for _, maybeCapStr := range candidates {
		maybeCap := rabbitCapability(strings.TrimSpace(maybeCapStr))
		enabled, present := allRabbitCapabilities[maybeCap]
		if enabled && present {
			result[maybeCap] = true
		}
	}
	return result
}

func isCapEnabled(config rabbitExporterConfig, cap rabbitCapability) bool {
	exists, enabled := config.RabbitCapabilities[cap]
	return exists && enabled
}
