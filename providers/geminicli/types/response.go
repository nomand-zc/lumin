package types

// GeminiCLIStreamResponse SSE 流中每个 data: 行的顶层结构
type GeminiCLIStreamResponse struct {
	Response *GeminiCLIResponseBody `json:"response,omitempty"`
}

// GeminiCLIResponseBody 响应体
type GeminiCLIResponseBody struct {
	Candidates    []GeminiCandidate `json:"candidates,omitempty"`
	UsageMetadata *GeminiUsage      `json:"usageMetadata,omitempty"`
	ModelVersion  string            `json:"modelVersion,omitempty"`
	ResponseID    string            `json:"responseId,omitempty"`
}

// GeminiCandidate 候选响应
type GeminiCandidate struct {
	Content      *GeminiCandidateContent `json:"content,omitempty"`
	FinishReason string                  `json:"finishReason,omitempty"`
}

// GeminiCandidateContent 候选内容
type GeminiCandidateContent struct {
	Parts []GeminiResponsePart `json:"parts,omitempty"`
	Role  string               `json:"role,omitempty"`
}

// GeminiResponsePart 响应中的一个部分
type GeminiResponsePart struct {
	Text             string                  `json:"text,omitempty"`
	Thought          bool                    `json:"thought,omitempty"`
	ThoughtSignature string                  `json:"thoughtSignature,omitempty"`
	FunctionCall     *GeminiResponseFuncCall `json:"functionCall,omitempty"`
}

// GeminiResponseFuncCall 响应中的函数调用
type GeminiResponseFuncCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

// GeminiUsage 用量元数据
type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
}

// GeminiCLIErrorResponse Gemini CLI 错误响应
type GeminiCLIErrorResponse struct {
	Error *GeminiCLIError `json:"error,omitempty"`
}

// GeminiCLIError 错误详情
type GeminiCLIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}
