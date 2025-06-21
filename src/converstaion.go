package main

var DEV_PROMPT string = `
You are a programming assistant. You might receive a bunch of error messages or general questions, your job is to answer them with any code snippets if necessary.
`

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIAPIRequest struct {
	Model    string           `json:"model"`
	Store    bool             `json:"store"`
	Messages []RequestMessage `json:"messages"`
}

// GPTResponse represents the structure of the API response
type GPTResponse struct {
	Choices           []Choice `json:"choices"`
	Created           float64  `json:"created"`
	ID                string   `json:"id"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	ServiceTier       string   `json:"service_tier"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Usage             Usage    `json:"usage"`
}

// Choice represents each choice returned in the response
type Choice struct {
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
	Logprobs     *string `json:"logprobs"`
	Message      Message `json:"message"`
}

// Message represents the message content in the response
type Message struct {
	Content string  `json:"content"`
	Refusal *string `json:"refusal"`
	Role    string  `json:"role"`
}

// Usage represents token usage details
type Usage struct {
	CompletionTokens        int          `json:"completion_tokens"`
	CompletionTokensDetails TokenDetails `json:"completion_tokens_details"`
	PromptTokens            int          `json:"prompt_tokens"`
	PromptTokensDetails     TokenDetails `json:"prompt_tokens_details"`
	TotalTokens             int          `json:"total_tokens"`
}

// TokenDetails represents details of token usage
type TokenDetails struct {
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	CachedTokens             int `json:"cached_tokens,omitempty"`
}

var chatMessages []RequestMessage = []RequestMessage{
	{
		Role:    "developer",
		Content: DEV_PROMPT,
	},
}

// Will send maximum of this many messages with rolling window
const MaxConvLen = 16

func GetConverstaionMessages(newMessage RequestMessage) []RequestMessage {
	// Keep the system message at index 0
	if len(chatMessages) > MaxConvLen {
		// Drop from index 1 (preserve system)
		// dropping idx 1 and 2 as idx 1 will be user prompt, and idx 2 will be assistant answer
		chatMessages = append(chatMessages[:1], chatMessages[3:]...)
	}

	chatMessages = append(chatMessages, newMessage)

	return chatMessages
}
