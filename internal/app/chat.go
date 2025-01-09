// File: /loam/internal/app/chat.go

package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	tea "github.com/charmbracelet/bubbletea"
)

// InferenceConfig represents the configuration for the inference.
type InferenceConfig struct {
	MaxNewTokens int `json:"max_new_tokens"`
}

// TextContent represents the structure of each content object.
type TextContent struct {
	Text string `json:"text"`
}

// Message represents a single message in the chat.
type Message struct {
	Role    string        `json:"role"`
	Content []TextContent `json:"content"`
}

// ChatRequest represents the request payload for AWS Bedrock's chat model.
type ChatRequest struct {
	InferenceConfig InferenceConfig `json:"inferenceConfig"`
	Messages        []Message       `json:"messages"`
}

// ChatResponse represents the response from AWS Bedrock's chat model.
type ChatResponse struct {
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"message"`
	} `json:"output"`
}

// ChatResponseMsg represents a successful response from the chat model.
type ChatResponseMsg struct {
	Response string
}

// ChatErrorMsg represents an error that occurred during chat invocation.
type ChatErrorMsg struct {
	Error error
}

// FoundationModelsMsg represents the list of foundation models retrieved.
type FoundationModelsMsg struct {
	Models []string
}

// ChatService encapsulates the AWS Bedrock Runtime and Bedrock clients.
type ChatService struct {
	BedrockClient      *bedrockruntime.Client
	BedrockModelClient *bedrock.Client
}

// NewChatService initializes the ChatService with the Bedrock Runtime and Bedrock clients.
// It accepts an optional profile name. If profileName is empty, the default profile is used.
func NewChatService(profileName string) (*ChatService, error) {
	var cfg aws.Config
	var err error

	if profileName != "" {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithSharedConfigProfile(profileName),
			config.WithRegion("us-east-1"),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion("us-east-1"),
		)
	}
	if err != nil {
		return nil, err
	}

	runtimeClient := bedrockruntime.NewFromConfig(cfg)
	modelClient := bedrock.NewFromConfig(cfg)

	return &ChatService{
		BedrockClient:      runtimeClient,
		BedrockModelClient: modelClient,
	}, nil
}

// SendChatCommand creates a Bubble Tea command that sends the prompt and context to Bedrock's Claude and returns the response.
func (cs *ChatService) SendChatCommand(prompt string, chatContext string) tea.Cmd {
	return func() tea.Msg {
		modelID := "amazon.nova-lite-v1:0"

		combinedText := fmt.Sprintf("%s\n\n%s", chatContext, prompt)

		userMessage := Message{
			Role: "user",
			Content: []TextContent{
				{Text: combinedText},
			},
		}

		requestPayload := ChatRequest{
			InferenceConfig: InferenceConfig{
				MaxNewTokens: 1000,
			},
			Messages: []Message{userMessage},
		}

		body, err := json.Marshal(requestPayload)
		if err != nil {
			return ChatErrorMsg{Error: err}
		}

		output, err := cs.BedrockClient.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(modelID),
			ContentType: aws.String("application/json"),
			Body:        body,
		})
		if err != nil {
			return ChatErrorMsg{Error: err}
		}

		var response ChatResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			return ChatErrorMsg{Error: err}
		}

		if len(response.Output.Message.Content) > 0 {
			assistantMessage := response.Output.Message.Content[0].Text
			return ChatResponseMsg{Response: assistantMessage}
		}

		return ChatErrorMsg{Error: fmt.Errorf("no assistant message found in the response")}
	}
}

// GetFoundationModels fetches the list of available foundation models.
func (cs *ChatService) GetFoundationModels() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		input := &bedrock.ListFoundationModelsInput{}

		output, err := cs.BedrockModelClient.ListFoundationModels(ctx, input)
		if err != nil {
			return ChatErrorMsg{Error: err}
		}

		if len(output.ModelSummaries) == 0 {
			return FoundationModelsMsg{Models: []string{"No foundation models available."}}
		}

		modelIDs := make([]string, len(output.ModelSummaries))
		for i, model := range output.ModelSummaries {
			modelIDs[i] = *model.ModelId
		}

		return FoundationModelsMsg{Models: modelIDs}
	}
}

// ProcessError is a helper function to handle errors from Bedrock invocation.
func ProcessError(err error, modelID string) {
}

var (
	chatServiceInstance *ChatService
	chatServiceErr      error
)

func init() {
	chatServiceInstance, chatServiceErr = NewChatService("")
}

// SendChat sends a chat message with context and returns a command to handle it.
func SendChat(prompt string, chatContext string) tea.Cmd {
	if chatServiceInstance == nil {
		return func() tea.Msg {
			return ChatErrorMsg{Error: chatServiceErr}
		}
	}
	return chatServiceInstance.SendChatCommand(prompt, chatContext)
}

// GetModels sends a command to fetch foundation models.
func GetModels() tea.Cmd {
	if chatServiceInstance == nil {
		return func() tea.Msg {
			return ChatErrorMsg{Error: chatServiceErr}
		}
	}
	return chatServiceInstance.GetFoundationModels()
}
