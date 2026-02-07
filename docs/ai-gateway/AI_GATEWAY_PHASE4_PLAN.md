# Phase 4: 平台化 - 详细设计文档

> **目标**: 将 Plumego AI 基础设施转化为完整的 Agent 开发平台
>
> **时间**: 12+ 周 | **状态**: 规划中 | **Go**: 1.24+

---

## 总览

Phase 4 在 Phase 1-3 的基础上，构建用户友好的平台层，使开发者能够快速创建、部署和管理 AI Agent 应用。

### 依赖关系

```
Phase 1 (Provider抽象) → Phase 2 (工具&提示) → Phase 3 (编排&分布式)
                                                    ↓
                                            Phase 4 (平台化)
```

---

## Sprint 1: Agent 市场 (3周)

### 目标
创建 Agent 市场，用户可以发布、发现、安装和使用预构建的 Agent 和工作流。

### 核心功能

#### 1.1 Agent 注册表 (Registry)

**数据结构**:
```go
// AgentMetadata 描述一个可重用的 Agent
type AgentMetadata struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Author      string            `json:"author"`
    Description string            `json:"description"`
    Category    AgentCategory     `json:"category"`
    Tags        []string          `json:"tags"`
    Provider    string            `json:"provider"` // claude, openai, etc.
    Model       string            `json:"model"`
    Prompt      PromptTemplate    `json:"prompt"`
    Tools       []ToolReference   `json:"tools"`
    Config      AgentConfig       `json:"config"`
    Dependencies []string         `json:"dependencies"`
    License     string            `json:"license"`
    Downloads   int64             `json:"downloads"`
    Rating      float64           `json:"rating"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type AgentCategory string
const (
    CategoryDataAnalysis   AgentCategory = "data_analysis"
    CategoryCodeGeneration AgentCategory = "code_generation"
    CategoryContentWriting AgentCategory = "content_writing"
    CategoryResearch       AgentCategory = "research"
    CategoryCustomerService AgentCategory = "customer_service"
    CategoryDevOps         AgentCategory = "devops"
)

type PromptTemplate struct {
    System      string                 `json:"system"`
    Variables   []VariableDefinition   `json:"variables"`
    Examples    []PromptExample        `json:"examples"`
}

type VariableDefinition struct {
    Name        string `json:"name"`
    Type        string `json:"type"` // string, number, boolean, array, object
    Required    bool   `json:"required"`
    Default     any    `json:"default,omitempty"`
    Description string `json:"description"`
}
```

**API 接口**:
```go
type AgentRegistry interface {
    // 发布 Agent
    Publish(ctx context.Context, metadata *AgentMetadata) error

    // 搜索 Agent
    Search(ctx context.Context, query SearchQuery) ([]*AgentMetadata, error)

    // 获取 Agent 详情
    Get(ctx context.Context, id string, version string) (*AgentMetadata, error)

    // 列出所有版本
    ListVersions(ctx context.Context, id string) ([]string, error)

    // 安装 Agent (下载到本地)
    Install(ctx context.Context, id string, version string) error

    // 评分和评论
    Rate(ctx context.Context, id string, rating float64, comment string) error

    // 获取依赖
    ResolveDependencies(ctx context.Context, id string) ([]*AgentMetadata, error)
}

type SearchQuery struct {
    Query      string
    Category   AgentCategory
    Tags       []string
    Provider   string
    SortBy     SortCriteria // downloads, rating, updated
    Limit      int
    Offset     int
}
```

**实现方案**:
```
ai/marketplace/
├── registry.go          # Agent 注册表核心
├── registry_local.go    # 本地文件系统实现
├── registry_remote.go   # 远程仓库实现
├── search.go            # 搜索引擎
├── installer.go         # Agent 安装器
├── validator.go         # Agent 验证
└── registry_test.go
```

#### 1.2 工作流市场 (Workflow Marketplace)

```go
type WorkflowTemplate struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Version     string              `json:"version"`
    Author      string              `json:"author"`
    Description string              `json:"description"`
    Category    WorkflowCategory    `json:"category"`
    Tags        []string            `json:"tags"`
    Steps       []StepTemplate      `json:"steps"`
    Variables   []VariableDefinition `json:"variables"`
    Outputs     []OutputDefinition  `json:"outputs"`
    Example     WorkflowExample     `json:"example"`
    License     string              `json:"license"`
}

type StepTemplate struct {
    Name        string         `json:"name"`
    Type        string         `json:"type"` // agent, parallel, conditional
    AgentRef    string         `json:"agent_ref"` // Reference to marketplace agent
    InputMap    map[string]any `json:"input_map"`
    OutputKey   string         `json:"output_key"`
}

type WorkflowExample struct {
    Input       map[string]any `json:"input"`
    ExpectedOutput map[string]any `json:"expected_output"`
    Description string         `json:"description"`
}
```

#### 1.3 包管理器 (Package Manager)

类似于 npm/pip，管理 Agent 依赖：

```go
type PackageManager struct {
    registry    AgentRegistry
    localStore  string // ~/.plumego/agents/
    lockFile    string // plumego.lock
}

// plumego.yaml - 项目配置文件
type ProjectManifest struct {
    Name         string            `yaml:"name"`
    Version      string            `yaml:"version"`
    Dependencies map[string]string `yaml:"dependencies"` // agent-id: version
    DevDependencies map[string]string `yaml:"dev_dependencies"`
}

// CLI 命令
// plumego agent install sentiment-analyzer@1.2.0
// plumego agent search "data analysis"
// plumego agent publish ./my-agent.yaml
// plumego workflow install data-pipeline@2.0.0
```

#### 1.4 验证和安全

```go
type AgentValidator struct{}

func (v *AgentValidator) Validate(agent *AgentMetadata) error {
    // 检查必填字段
    // 验证 prompt 语法
    // 检查工具引用是否存在
    // 验证配置参数
    // 扫描安全问题（prompt injection, etc.）
}

type SecurityScanner struct{}

func (s *SecurityScanner) ScanAgent(agent *AgentMetadata) (*SecurityReport, error) {
    // 检测恶意 prompt
    // 验证工具权限
    // 检查敏感数据泄露
}
```

---

## Sprint 2: 向量嵌入缓存增强 (2周)

### 目标
扩展 Phase 3 Sprint 3 的语义缓存，支持自定义嵌入模型、多级缓存和向量数据库集成。

### 核心功能

#### 2.1 多嵌入模型支持

```go
// 扩展现有 EmbeddingGenerator 接口
type EmbeddingProvider interface {
    Name() string
    Generate(ctx context.Context, text string) (*Embedding, error)
    GenerateBatch(ctx context.Context, texts []string) ([]*Embedding, error)
    Dimensions() int
    Model() string
    CostPerToken() float64 // 成本计算
}

// 支持的嵌入模型
type OpenAIEmbeddingProvider struct {
    apiKey string
    model  string // text-embedding-3-small, text-embedding-3-large
}

type VoyageAIEmbeddingProvider struct {
    apiKey string
    model  string // voyage-2, voyage-code-2
}

type LocalEmbeddingProvider struct {
    modelPath string // 本地 ONNX 模型路径
    runtime   *onnxruntime.Session
}

type CoherEmbeddingProvider struct {
    apiKey string
    model  string // embed-english-v3.0, embed-multilingual-v3.0
}
```

#### 2.2 向量数据库集成

```go
// 扩展 VectorStore 支持多种后端
type VectorStoreBackend string
const (
    BackendMemory     VectorStoreBackend = "memory"
    BackendPinecone   VectorStoreBackend = "pinecone"
    BackendWeaviate   VectorStoreBackend = "weaviate"
    BackendQdrant     VectorStoreBackend = "qdrant"
    BackendMilvus     VectorStoreBackend = "milvus"
    BackendChroma     VectorStoreBackend = "chroma"
)

type VectorStoreConfig struct {
    Backend    VectorStoreBackend
    Endpoint   string
    APIKey     string
    Index      string
    Dimensions int
}

// Pinecone 实现
type PineconeVectorStore struct {
    client    *pinecone.Client
    index     string
    namespace string
}

func (p *PineconeVectorStore) Add(ctx context.Context, entry *VectorEntry) error
func (p *PineconeVectorStore) Search(ctx context.Context, query *Embedding, topK int, threshold float64) ([]*SimilarityResult, error)
func (p *PineconeVectorStore) Delete(ctx context.Context, hash string) error
```

#### 2.3 多级缓存策略

```go
type TieredCacheStrategy struct {
    L1 VectorStore // Memory (fast, small)
    L2 VectorStore // Redis (medium, larger)
    L3 VectorStore // Vector DB (slow, unlimited)

    L1MaxSize int
    L2MaxSize int
}

func (t *TieredCacheStrategy) Get(ctx context.Context, query *Embedding) (*CacheEntry, error) {
    // Try L1 first
    if result, err := t.L1.Search(ctx, query, 1, 0.85); err == nil && len(result) > 0 {
        return result[0].Entry, nil
    }

    // Try L2
    if result, err := t.L2.Search(ctx, query, 1, 0.85); err == nil && len(result) > 0 {
        // Promote to L1
        t.L1.Add(ctx, result[0].Entry)
        return result[0].Entry, nil
    }

    // Try L3
    if result, err := t.L3.Search(ctx, query, 1, 0.85); err == nil && len(result) > 0 {
        // Promote to L2 and L1
        t.L2.Add(ctx, result[0].Entry)
        t.L1.Add(ctx, result[0].Entry)
        return result[0].Entry, nil
    }

    return nil, ErrCacheMiss
}
```

#### 2.4 缓存预热和管理

```go
type CacheWarmer struct {
    cache      *SemanticCache
    embeddings EmbeddingProvider
}

// 从历史查询预热缓存
func (w *CacheWarmer) WarmFromHistory(ctx context.Context, queries []string) error

// 从文档库预热缓存
func (w *CacheWarmer) WarmFromDocuments(ctx context.Context, docs []Document) error

type CacheManager struct {
    cache *SemanticCache
}

// 缓存统计
func (m *CacheManager) Stats() CacheStats

// 缓存清理策略
func (m *CacheManager) Cleanup(policy CleanupPolicy) error

// 导出/导入缓存
func (m *CacheManager) Export(path string) error
func (m *CacheManager) Import(path string) error
```

---

## Sprint 3: 多模态支持 (2-3周)

### 目标
扩展 Provider 抽象以支持图片、音频、视频等多模态输入输出。

### 核心功能

#### 3.1 多模态消息类型

```go
// 扩展现有 Message 结构
type Message struct {
    Role     Role
    Content  []ContentBlock // 支持多个内容块
    Name     string
    ToolCallID string
}

type ContentBlock struct {
    Type     ContentType
    Text     string
    Image    *ImageContent
    Audio    *AudioContent
    Video    *VideoContent
    Document *DocumentContent
    ToolUse  *ToolUse
    ToolResult *ToolResult
}

type ContentType string
const (
    ContentTypeText     ContentType = "text"
    ContentTypeImage    ContentType = "image"
    ContentTypeAudio    ContentType = "audio"
    ContentTypeVideo    ContentType = "video"
    ContentTypeDocument ContentType = "document"
    ContentTypeToolUse  ContentType = "tool_use"
    ContentTypeToolResult ContentType = "tool_result"
)

type ImageContent struct {
    Format ImageFormat  `json:"format"` // jpeg, png, webp, gif
    Source ImageSource  `json:"source"`
    Data   []byte       `json:"data,omitempty"`   // Base64 encoded
    URL    string       `json:"url,omitempty"`    // Or URL
    Width  int          `json:"width,omitempty"`
    Height int          `json:"height,omitempty"`
}

type ImageFormat string
const (
    ImageFormatJPEG ImageFormat = "jpeg"
    ImageFormatPNG  ImageFormat = "png"
    ImageFormatWebP ImageFormat = "webp"
    ImageFormatGIF  ImageFormat = "gif"
)

type ImageSource string
const (
    ImageSourceBase64 ImageSource = "base64"
    ImageSourceURL    ImageSource = "url"
)

type AudioContent struct {
    Format   AudioFormat `json:"format"` // mp3, wav, ogg
    Data     []byte      `json:"data,omitempty"`
    URL      string      `json:"url,omitempty"`
    Duration float64     `json:"duration,omitempty"` // seconds
    Transcript string    `json:"transcript,omitempty"`
}

type AudioFormat string
const (
    AudioFormatMP3 AudioFormat = "mp3"
    AudioFormatWAV AudioFormat = "wav"
    AudioFormatOGG AudioFormat = "ogg"
)

type VideoContent struct {
    Format    VideoFormat `json:"format"` // mp4, webm
    Data      []byte      `json:"data,omitempty"`
    URL       string      `json:"url,omitempty"`
    Duration  float64     `json:"duration,omitempty"`
    Width     int         `json:"width,omitempty"`
    Height    int         `json:"height,omitempty"`
    FrameRate float64     `json:"frame_rate,omitempty"`
}

type VideoFormat string
const (
    VideoFormatMP4  VideoFormat = "mp4"
    VideoFormatWebM VideoFormat = "webm"
)

type DocumentContent struct {
    Format   DocumentFormat `json:"format"` // pdf, docx, txt
    Data     []byte         `json:"data,omitempty"`
    URL      string         `json:"url,omitempty"`
    PageCount int           `json:"page_count,omitempty"`
    Text     string         `json:"text,omitempty"` // Extracted text
}

type DocumentFormat string
const (
    DocumentFormatPDF  DocumentFormat = "pdf"
    DocumentFormatDOCX DocumentFormat = "docx"
    DocumentFormatTXT  DocumentFormat = "txt"
)
```

#### 3.2 多模态 Provider 接口

```go
type MultimodalProvider interface {
    provider.Provider

    // 支持的模态
    SupportedModalities() []ContentType

    // 检查模态支持
    SupportsModality(modality ContentType) bool

    // 多模态完成
    CompleteMultimodal(ctx context.Context, req *MultimodalCompletionRequest) (*provider.CompletionResponse, error)

    // 图片生成（如 DALL-E）
    GenerateImage(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error)

    // 语音合成（如 TTS）
    SynthesizeSpeech(ctx context.Context, req *SpeechSynthesisRequest) (*AudioContent, error)

    // 语音识别（如 Whisper）
    TranscribeAudio(ctx context.Context, audio *AudioContent) (*TranscriptionResult, error)

    // 视频分析
    AnalyzeVideo(ctx context.Context, video *VideoContent) (*VideoAnalysisResult, error)
}

type MultimodalCompletionRequest struct {
    Model       string
    Messages    []Message // 可以包含多模态内容
    MaxTokens   int
    Temperature float64
}

type ImageGenerationRequest struct {
    Prompt  string
    Model   string // dall-e-3, stable-diffusion, etc.
    Size    string // 1024x1024, 512x512, etc.
    Quality string // standard, hd
    Style   string // vivid, natural
    N       int    // Number of images
}

type ImageGenerationResponse struct {
    Images    []*ImageContent
    Revised   string // Revised prompt
    CreatedAt time.Time
}

type SpeechSynthesisRequest struct {
    Text   string
    Model  string // tts-1, tts-1-hd
    Voice  string // alloy, echo, fable, onyx, nova, shimmer
    Speed  float64
    Format AudioFormat
}

type TranscriptionResult struct {
    Text     string
    Language string
    Duration float64
    Segments []TranscriptionSegment
}

type TranscriptionSegment struct {
    Text      string
    Start     float64
    End       float64
    Confidence float64
}

type VideoAnalysisResult struct {
    Description string
    Objects     []DetectedObject
    Scenes      []SceneDescription
    Transcript  *TranscriptionResult
}

type DetectedObject struct {
    Name       string
    Confidence float64
    BoundingBox BoundingBox
    Timestamp   float64
}

type BoundingBox struct {
    X      int
    Y      int
    Width  int
    Height int
}

type SceneDescription struct {
    Start       float64
    End         float64
    Description string
    KeyFrames   []*ImageContent
}
```

#### 3.3 多模态工具

```go
// 图片处理工具
type ImageProcessingTool struct {
    name string
}

func (t *ImageProcessingTool) Resize(img *ImageContent, width, height int) (*ImageContent, error)
func (t *ImageProcessingTool) Crop(img *ImageContent, x, y, width, height int) (*ImageContent, error)
func (t *ImageProcessingTool) Filter(img *ImageContent, filter string) (*ImageContent, error)
func (t *ImageProcessingTool) OCR(img *ImageContent) (string, error)

// 视频处理工具
type VideoProcessingTool struct {
    name string
}

func (t *VideoProcessingTool) ExtractFrames(video *VideoContent, fps float64) ([]*ImageContent, error)
func (t *VideoProcessingTool) ExtractAudio(video *VideoContent) (*AudioContent, error)
func (t *VideoProcessingTool) Trim(video *VideoContent, start, end float64) (*VideoContent, error)

// 文档处理工具
type DocumentProcessingTool struct {
    name string
}

func (t *DocumentProcessingTool) ExtractText(doc *DocumentContent) (string, error)
func (t *DocumentProcessingTool) ExtractImages(doc *DocumentContent) ([]*ImageContent, error)
func (t *DocumentProcessingTool) ConvertToPDF(doc *DocumentContent) (*DocumentContent, error)
```

#### 3.4 多模态 Agent 示例

```go
type VisionAgent struct {
    provider MultimodalProvider
    tools    []Tool
}

// 图片描述 Agent
func NewImageDescriptionAgent(provider MultimodalProvider) *VisionAgent {
    return &VisionAgent{
        provider: provider,
    }
}

func (a *VisionAgent) DescribeImage(ctx context.Context, image *ImageContent, query string) (string, error) {
    req := &MultimodalCompletionRequest{
        Model: "gpt-4-vision-preview",
        Messages: []Message{
            {
                Role: RoleUser,
                Content: []ContentBlock{
                    {Type: ContentTypeImage, Image: image},
                    {Type: ContentTypeText, Text: query},
                },
            },
        },
    }

    resp, err := a.provider.CompleteMultimodal(ctx, req)
    if err != nil {
        return "", err
    }

    return resp.GetText(), nil
}

// 视频摘要 Agent
func NewVideoSummaryAgent(provider MultimodalProvider) *VisionAgent {
    return &VisionAgent{
        provider: provider,
        tools: []Tool{
            &VideoProcessingTool{},
        },
    }
}

func (a *VisionAgent) SummarizeVideo(ctx context.Context, video *VideoContent) (string, error) {
    // 1. 提取关键帧
    frames, _ := a.tools[0].(*VideoProcessingTool).ExtractFrames(video, 1.0)

    // 2. 分析每一帧
    descriptions := []string{}
    for _, frame := range frames {
        desc, _ := a.DescribeImage(ctx, frame, "Describe this scene briefly")
        descriptions = append(descriptions, desc)
    }

    // 3. 生成摘要
    req := &provider.CompletionRequest{
        Model: "gpt-4",
        Messages: []provider.Message{
            provider.NewTextMessage(provider.RoleUser,
                fmt.Sprintf("Summarize this video based on these scene descriptions: %s",
                    strings.Join(descriptions, "\n"))),
        },
    }

    resp, _ := a.provider.Complete(ctx, req)
    return resp.GetText(), nil
}
```

---

## Sprint 4: 自定义工具 SDK (2周)

### 目标
简化自定义工具的创建、测试、发布和集成流程。

### 核心功能

#### 4.1 工具开发框架

```go
// 简化的工具定义接口
type SimpleTool interface {
    // 工具元数据
    Name() string
    Description() string

    // 参数定义
    Parameters() ParameterSchema

    // 执行逻辑
    Execute(ctx context.Context, params map[string]any) (any, error)

    // 可选：异步执行
    ExecuteAsync(ctx context.Context, params map[string]any) (<-chan any, error)
}

type ParameterSchema struct {
    Type       string                    `json:"type"` // object
    Properties map[string]PropertySchema `json:"properties"`
    Required   []string                  `json:"required"`
}

type PropertySchema struct {
    Type        string   `json:"type"` // string, number, boolean, array, object
    Description string   `json:"description"`
    Enum        []any    `json:"enum,omitempty"`
    Default     any      `json:"default,omitempty"`
    Minimum     *float64 `json:"minimum,omitempty"`
    Maximum     *float64 `json:"maximum,omitempty"`
}

// 工具构建器 - 声明式 API
type ToolBuilder struct {
    tool *toolImpl
}

func NewTool(name, description string) *ToolBuilder {
    return &ToolBuilder{
        tool: &toolImpl{
            name:        name,
            description: description,
            params:      ParameterSchema{Type: "object", Properties: make(map[string]PropertySchema)},
        },
    }
}

func (b *ToolBuilder) AddStringParam(name, description string, required bool) *ToolBuilder {
    b.tool.params.Properties[name] = PropertySchema{
        Type:        "string",
        Description: description,
    }
    if required {
        b.tool.params.Required = append(b.tool.params.Required, name)
    }
    return b
}

func (b *ToolBuilder) AddNumberParam(name, description string, min, max *float64, required bool) *ToolBuilder {
    b.tool.params.Properties[name] = PropertySchema{
        Type:        "number",
        Description: description,
        Minimum:     min,
        Maximum:     max,
    }
    if required {
        b.tool.params.Required = append(b.tool.params.Required, name)
    }
    return b
}

func (b *ToolBuilder) SetHandler(handler func(ctx context.Context, params map[string]any) (any, error)) *ToolBuilder {
    b.tool.handler = handler
    return b
}

func (b *ToolBuilder) Build() SimpleTool {
    return b.tool
}

// 使用示例
weatherTool := NewTool("get_weather", "Get current weather for a location").
    AddStringParam("location", "City name or coordinates", true).
    AddStringParam("units", "Temperature units (celsius/fahrenheit)", false).
    SetHandler(func(ctx context.Context, params map[string]any) (any, error) {
        location := params["location"].(string)
        units := params["units"].(string)

        // 调用天气 API
        weather, err := getWeatherFromAPI(location, units)
        return weather, err
    }).
    Build()
```

#### 4.2 工具验证和测试框架

```go
type ToolTester struct {
    tool SimpleTool
}

func NewToolTester(tool SimpleTool) *ToolTester {
    return &ToolTester{tool: tool}
}

// 验证工具定义
func (t *ToolTester) Validate() error {
    // 检查必填字段
    // 验证参数 schema
    // 确保 handler 不为 nil
}

// 单元测试
func (t *ToolTester) Test(testCases []ToolTestCase) (*ToolTestReport, error) {
    report := &ToolTestReport{}

    for _, tc := range testCases {
        result, err := t.tool.Execute(context.Background(), tc.Input)

        if err != nil {
            report.Failures = append(report.Failures, ToolTestFailure{
                TestCase: tc,
                Error:    err,
            })
        } else if !reflect.DeepEqual(result, tc.Expected) {
            report.Failures = append(report.Failures, ToolTestFailure{
                TestCase: tc,
                Actual:   result,
            })
        } else {
            report.Passed++
        }
    }

    return report, nil
}

type ToolTestCase struct {
    Name     string
    Input    map[string]any
    Expected any
    ShouldFail bool
}

type ToolTestReport struct {
    Passed   int
    Failed   int
    Failures []ToolTestFailure
}

type ToolTestFailure struct {
    TestCase ToolTestCase
    Actual   any
    Error    error
}

// 工具配置文件 tool.yaml
type ToolManifest struct {
    Name        string              `yaml:"name"`
    Version     string              `yaml:"version"`
    Description string              `yaml:"description"`
    Author      string              `yaml:"author"`
    License     string              `yaml:"license"`
    Parameters  ParameterSchema     `yaml:"parameters"`
    Handler     string              `yaml:"handler"` // Path to Go file
    Tests       []ToolTestCase      `yaml:"tests"`
    Dependencies map[string]string  `yaml:"dependencies"`
}
```

#### 4.3 工具打包和发布

```bash
# CLI 工具
plumego tool create my-weather-tool
plumego tool test
plumego tool build
plumego tool publish

# 项目结构
my-weather-tool/
├── tool.yaml          # 工具配置
├── handler.go         # 工具实现
├── handler_test.go    # 单元测试
├── README.md          # 文档
└── examples/          # 使用示例
    └── example.go
```

```go
type ToolPackager struct {
    manifest *ToolManifest
}

// 打包工具为可分发格式
func (p *ToolPackager) Package() ([]byte, error) {
    // 1. 验证工具
    // 2. 运行测试
    // 3. 编译 Go 代码为插件（或打包源码）
    // 4. 生成 tarball
}

// 发布到工具市场
func (p *ToolPackager) Publish(registry ToolRegistry) error {
    // 1. 打包工具
    // 2. 上传到注册表
    // 3. 更新索引
}
```

#### 4.4 工具插件系统

```go
// 支持动态加载工具
type PluginLoader struct {
    pluginDir string
}

func (l *PluginLoader) LoadTool(name string) (SimpleTool, error) {
    // 使用 Go plugin 系统或 WASM
    pluginPath := filepath.Join(l.pluginDir, name+".so")

    plugin, err := plugin.Open(pluginPath)
    if err != nil {
        return nil, err
    }

    symbol, err := plugin.Lookup("Tool")
    if err != nil {
        return nil, err
    }

    tool, ok := symbol.(SimpleTool)
    if !ok {
        return nil, fmt.Errorf("invalid tool plugin")
    }

    return tool, nil
}

// 或使用 WASM
type WASMTool struct {
    runtime *wazero.Runtime
    module  wazero.CompiledModule
}

func LoadWASMTool(wasmPath string) (*WASMTool, error) {
    // 加载 WASM 模块
    // 实现跨语言工具支持（Python, JavaScript, Rust）
}
```

---

## Sprint 5: 低代码编排器 (3-4周)

### 目标
提供可视化界面和 DSL，让非开发者也能创建 AI 工作流。

### 核心功能

#### 5.1 工作流 DSL

```yaml
# workflow.yaml - 声明式工作流定义
name: customer-support-pipeline
version: 1.0.0
description: Automated customer support workflow

variables:
  - name: customer_email
    type: string
    required: true
  - name: issue_category
    type: string
    default: "general"

steps:
  - name: classify
    type: agent
    agent: sentiment-classifier@1.0.0
    input: "{{ variables.customer_email }}"
    output: sentiment

  - name: route
    type: conditional
    condition: "{{ steps.classify.output.sentiment == 'negative' }}"
    then:
      - name: escalate
        type: agent
        agent: human-escalator@1.0.0
        input:
          email: "{{ variables.customer_email }}"
          priority: "high"
    else:
      - name: auto_respond
        type: agent
        agent: response-generator@2.1.0
        input:
          email: "{{ variables.customer_email }}"
          category: "{{ variables.issue_category }}"
        output: response

  - name: parallel_tasks
    type: parallel
    agents:
      - name: log_ticket
        agent: ticket-logger@1.0.0
        input: "{{ variables.customer_email }}"
      - name: update_crm
        agent: crm-updater@1.5.0
        input:
          email: "{{ variables.customer_email }}"
          status: "processed"

outputs:
  - name: response
    value: "{{ steps.auto_respond.output }}"
  - name: ticket_id
    value: "{{ steps.log_ticket.output.id }}"
```

```go
type WorkflowDSL struct {
    Name        string                  `yaml:"name"`
    Version     string                  `yaml:"version"`
    Description string                  `yaml:"description"`
    Variables   []VariableDefinition    `yaml:"variables"`
    Steps       []DSLStep               `yaml:"steps"`
    Outputs     []OutputDefinition      `yaml:"outputs"`
    OnError     *ErrorHandler           `yaml:"on_error,omitempty"`
}

type DSLStep struct {
    Name      string         `yaml:"name"`
    Type      string         `yaml:"type"` // agent, conditional, parallel, loop
    Agent     string         `yaml:"agent,omitempty"` // agent-id@version
    Input     any            `yaml:"input"`
    Output    string         `yaml:"output,omitempty"`
    Condition string         `yaml:"condition,omitempty"`
    Then      []DSLStep      `yaml:"then,omitempty"`
    Else      []DSLStep      `yaml:"else,omitempty"`
    Agents    []ParallelDSLAgent `yaml:"agents,omitempty"`
    Retry     *RetryConfig   `yaml:"retry,omitempty"`
}

type ParallelDSLAgent struct {
    Name   string `yaml:"name"`
    Agent  string `yaml:"agent"`
    Input  any    `yaml:"input"`
    Output string `yaml:"output,omitempty"`
}

type OutputDefinition struct {
    Name  string `yaml:"name"`
    Value string `yaml:"value"` // Template expression
}

type ErrorHandler struct {
    RetryCount int      `yaml:"retry_count"`
    Fallback   []DSLStep `yaml:"fallback"`
    Notify     []string `yaml:"notify"` // Email addresses
}

// DSL 解析器
type DSLParser struct {
    registry AgentRegistry
}

func (p *DSLParser) Parse(yamlContent []byte) (*orchestration.Workflow, error) {
    var dsl WorkflowDSL
    if err := yaml.Unmarshal(yamlContent, &dsl); err != nil {
        return nil, err
    }

    // 验证 DSL
    if err := p.validate(&dsl); err != nil {
        return nil, err
    }

    // 转换为 orchestration.Workflow
    workflow := orchestration.NewWorkflow(dsl.Name, dsl.Name, dsl.Description)

    for _, step := range dsl.Steps {
        orchStep, err := p.convertStep(&step)
        if err != nil {
            return nil, err
        }
        workflow.AddStep(orchStep)
    }

    return workflow, nil
}

// 模板引擎
type TemplateEngine struct{}

func (e *TemplateEngine) Evaluate(template string, context map[string]any) (any, error) {
    // 支持 {{ variables.name }}, {{ steps.step1.output }} 等语法
    // 使用 text/template 或 expr 库
}
```

#### 5.2 可视化编排器 Web UI

```
frontend/
├── workbench/
│   ├── canvas/           # 工作流画布（拖拽）
│   │   ├── FlowCanvas.tsx
│   │   ├── NodePalette.tsx
│   │   ├── AgentNode.tsx
│   │   ├── ConditionalNode.tsx
│   │   └── ParallelNode.tsx
│   ├── inspector/        # 节点属性面板
│   │   ├── NodeInspector.tsx
│   │   └── VariableEditor.tsx
│   ├── execution/        # 执行监控
│   │   ├── ExecutionView.tsx
│   │   ├── StepStatus.tsx
│   │   └── LogViewer.tsx
│   └── marketplace/      # Agent 市场集成
│       ├── AgentSearch.tsx
│       └── AgentDetails.tsx
```

**技术栈**:
- React Flow / xyflow - 工作流可视化
- Monaco Editor - 代码编辑器
- TailwindCSS - UI 样式
- React Query - 数据管理
- Zustand - 状态管理

**功能**:
1. 拖拽式工作流设计
2. 实时 DSL 预览
3. 执行监控和调试
4. 版本控制和协作
5. 模板库

#### 5.3 工作流模拟器

```go
type WorkflowSimulator struct {
    workflow *orchestration.Workflow
}

// 模拟执行（不调用真实 API）
func (s *WorkflowSimulator) Simulate(ctx context.Context, input map[string]any) (*SimulationResult, error) {
    result := &SimulationResult{
        Steps: make([]StepSimulation, 0),
    }

    for i, step := range s.workflow.Steps {
        sim := StepSimulation{
            StepIndex: i,
            StepName:  step.Name(),
            Input:     s.getStepInput(step, input),
        }

        // 使用 mock 数据
        sim.Output = s.getMockOutput(step)
        sim.Duration = time.Duration(rand.Intn(1000)) * time.Millisecond

        result.Steps = append(result.Steps, sim)
    }

    return result, nil
}

type SimulationResult struct {
    Steps         []StepSimulation
    TotalDuration time.Duration
    EstimatedCost float64
}

type StepSimulation struct {
    StepIndex int
    StepName  string
    Input     any
    Output    any
    Duration  time.Duration
}
```

#### 5.4 工作流调试器

```go
type WorkflowDebugger struct {
    workflow  *orchestration.Workflow
    engine    *orchestration.Engine
    breakpoints []Breakpoint
}

type Breakpoint struct {
    StepIndex int
    Condition string // Optional condition
}

func (d *WorkflowDebugger) Debug(ctx context.Context, input map[string]any) error {
    // 启用调试模式
    for i, step := range d.workflow.Steps {
        // 检查断点
        if d.hasBreakpoint(i) {
            // 暂停执行
            fmt.Printf("Breakpoint at step %d: %s\n", i, step.Name())
            fmt.Printf("State: %+v\n", d.workflow.State)

            // 等待用户输入
            // - continue
            // - step over
            // - inspect variables
            // - modify state
        }

        // 执行步骤
        result, err := step.Execute(ctx, d.workflow)
        if err != nil {
            fmt.Printf("Error at step %d: %v\n", i, err)
            return err
        }

        fmt.Printf("Step %d completed: %+v\n", i, result)
    }

    return nil
}
```

---

## 实现路线图

### 时间线

```
Week 1-3:  Sprint 1 - Agent 市场
Week 4-5:  Sprint 2 - 向量嵌入缓存增强
Week 6-8:  Sprint 3 - 多模态支持
Week 9-10: Sprint 4 - 自定义工具 SDK
Week 11-14: Sprint 5 - 低代码编排器
```

### 依赖关系

```
Sprint 1 (Agent市场) ←─┐
                        ├→ Sprint 5 (低代码编排器)
Sprint 4 (工具SDK)   ←─┘

Sprint 2 (缓存) ← Phase 3 Sprint 3 (语义缓存)

Sprint 3 (多模态) ← Phase 1 (Provider抽象)
```

---

## 目录结构

```
plumego/
├── ai/
│   ├── marketplace/         # Sprint 1
│   │   ├── registry/
│   │   ├── installer/
│   │   ├── validator/
│   │   └── search/
│   ├── semanticcache/       # Sprint 2 (扩展)
│   │   ├── embeddings/
│   │   ├── vectordb/
│   │   └── tiered/
│   ├── multimodal/          # Sprint 3
│   │   ├── provider/
│   │   ├── content/
│   │   └── tools/
│   ├── toolsdk/             # Sprint 4
│   │   ├── builder/
│   │   ├── tester/
│   │   ├── packager/
│   │   └── plugin/
│   └── workbench/           # Sprint 5
│       ├── dsl/
│       ├── parser/
│       ├── simulator/
│       └── debugger/
├── frontend/                # Sprint 5 Web UI
│   └── workbench/
└── examples/
    └── phase4/
```

---

## 成功指标

### Sprint 1: Agent 市场
- ✅ 至少 10 个预构建 Agent 可用
- ✅ Agent 搜索延迟 < 100ms
- ✅ 依赖解析正确率 100%

### Sprint 2: 向量缓存
- ✅ 支持 3+ 种向量数据库
- ✅ L1 缓存命中率 > 40%
- ✅ 查询延迟 < 50ms (P99)

### Sprint 3: 多模态
- ✅ 支持 4+ 种内容类型
- ✅ 图片处理延迟 < 2s
- ✅ 视频分析准确率 > 85%

### Sprint 4: 工具 SDK
- ✅ 工具创建时间 < 30 分钟
- ✅ 测试覆盖率 > 80%
- ✅ CLI 命令成功率 > 95%

### Sprint 5: 低代码编排器
- ✅ 支持 10+ 种工作流模板
- ✅ DSL 解析错误率 < 1%
- ✅ 工作流创建时间 < 10 分钟
- ✅ UI 响应时间 < 500ms

---

## 总结

Phase 4 将 Plumego 从基础设施转变为完整的 Agent 开发平台：

1. **Agent 市场** - 可重用组件生态
2. **向量缓存** - 性能优化
3. **多模态** - 扩展能力边界
4. **工具 SDK** - 简化开发流程
5. **低代码编排器** - 降低使用门槛

预期产出：
- **~8,000 LOC** 平台代码
- **~3,000 LOC** 前端代码
- **100+ 测试**
- **完整文档**
- **示例工作流库**
