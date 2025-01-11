package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bmquinn/loam-iiif/internal/app"
	"github.com/bmquinn/loam-iiif/internal/iiif"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Define command-line flags
	manifestURL := flag.String("manifest", "", "IIIF manifest URL")
	prompt := flag.String("prompt", "", "Prompt to send to the model")
	profile := flag.String("profile", "", "AWS profile to use (optional)")
	flag.Parse()

	// Check if both --manifest and --prompt are provided
	if *manifestURL != "" && *prompt != "" {
		// Run in command-line mode
		response, err := runCommandLine(*manifestURL, *prompt, *profile)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println(response)
		os.Exit(0)
	}

	// Otherwise, launch the TUI
	p := tea.NewProgram(app.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// runCommandLine handles the command-line operation
func runCommandLine(manifestURL, prompt, profile string) (string, error) {
	// Step 1: Fetch the IIIF manifest
	data, err := iiif.FetchDataSync(manifestURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// Step 2: Parse the IIIF manifest
	items := iiif.ParseData(data)
	if len(items) == 0 {
		return "", fmt.Errorf("no items found in the manifest")
	}

	// Step 3: Extract context from the parsed items
	var contextBuilder strings.Builder
	for _, item := range items {
		contextBuilder.WriteString(fmt.Sprintf("Title: %s\nURL: %s\n\n", item.Title, item.URL))
	}
	context := contextBuilder.String()

	// Step 4: Initialize the ChatService
	chatService, err := app.NewChatService(profile)
	if err != nil {
		return "", fmt.Errorf("failed to initialize chat service: %w", err)
	}

	// Step 5: Send the prompt and get the response
	response, err := chatService.SendChatSync(prompt, context)
	if err != nil {
		return "", fmt.Errorf("failed to send prompt: %w", err)
	}

	return response, nil
}
