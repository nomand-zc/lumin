package types

// GeminiCLIRequest 是发送到 Gemini CLI API 的顶层请求结构
type GeminiCLIRequest struct {
	Model              string          `json:"model"`
	Project            string          `json:"project,omitempty"`
	UserPromptID       string          `json:"user_prompt_id,omitempty"`
	Request            *GeminiCLIInner `json:"request"`
	EnabledCreditTypes []string        `json:"enabled_credit_types,omitempty"`
}

// GeminiCLIInner 是嵌套的 request 部分
type GeminiCLIInner struct {
	Contents          []GeminiContent         `json:"contents"`
	SystemInstruction *GeminiContent          `json:"systemInstruction,omitempty"`
	CachedContent     string                  `json:"cachedContent,omitempty"`
	Tools             []GeminiTool            `json:"tools,omitempty"`
	ToolConfig        *GeminiToolConfig       `json:"toolConfig,omitempty"`
	Labels            map[string]string       `json:"labels,omitempty"`
	GenerationConfig  *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings    []map[string]string     `json:"safetySettings"`
	SessionID         string                  `json:"session_id,omitempty"`
}

// GeminiContent 代表一条消息
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart 代表消息中的一个部分（请求侧）
type GeminiPart struct {
	Text             string              `json:"text,omitempty"`
	Thought          *bool               `json:"thought,omitempty"`
	ThoughtSignature string              `json:"thoughtSignature,omitempty"`
	FunctionCall     *GeminiFunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFuncResponse `json:"functionResponse,omitempty"`
	InlineData       *GeminiInlineData   `json:"inlineData,omitempty"`
}

// GeminiFunctionCall 代表函数调用
type GeminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

// GeminiFuncResponse 代表函数响应
type GeminiFuncResponse struct {
	Name     string              `json:"name"`
	Response GeminiFuncRespValue `json:"response"`
}

// GeminiFuncRespValue 代表函数响应的值
type GeminiFuncRespValue struct {
	Result string `json:"result"`
}

// GeminiInlineData 代表内联数据（如图片）
type GeminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GeminiTool 代表工具定义
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDecl `json:"functionDeclarations,omitempty"`
}

// GeminiFunctionDecl 代表函数声明
type GeminiFunctionDecl struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	ParametersJsonSchema any    `json:"parametersJsonSchema,omitempty"`
}

// GeminiToolConfig 工具配置
type GeminiToolConfig struct {
	FunctionCallingConfig *GeminiFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// GeminiFunctionCallingConfig 函数调用配置
type GeminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

// GeminiGenerationConfig 生成配置
type GeminiGenerationConfig struct {
	Temperature      *float64              `json:"temperature,omitempty"`
	TopP             *float64              `json:"topP,omitempty"`
	TopK             *float64              `json:"topK,omitempty"`
	MaxOutputTokens  *int                  `json:"maxOutputTokens,omitempty"`
	StopSequences    []string              `json:"stopSequences,omitempty"`
	PresencePenalty  *float64              `json:"presencePenalty,omitempty"`
	FrequencyPenalty *float64              `json:"frequencyPenalty,omitempty"`
	ThinkingConfig   *GeminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

// GeminiThinkingConfig thinking 配置
type GeminiThinkingConfig struct {
	ThinkingBudget  *int   `json:"thinkingBudget,omitempty"`
	ThinkingLevel   string `json:"thinkingLevel,omitempty"`
	IncludeThoughts *bool  `json:"includeThoughts,omitempty"`
}
