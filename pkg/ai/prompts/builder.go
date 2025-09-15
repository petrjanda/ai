package prompts

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/getsynq/ai/pkg/ai"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// Block represents a content block that can be rendered to markdown
type Block struct {
	content string
	title   string
	level   int
}

// BlockOption is a function that configures a block
type BlockOption func(*Block)

// WithTitle sets the title for a block
func WithTitle(title string) BlockOption {
	return func(b *Block) {
		b.title = title
	}
}

// WithLevel sets the heading level for a block (1-6)
func WithLevel(level int) BlockOption {
	return func(b *Block) {
		if level >= 1 && level <= 6 {
			b.level = level
		}
	}
}

// PromptBuilder allows building prompts with composable blocks
type PromptBuilder struct {
	blocks []*Block
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		blocks: make([]*Block, 0),
	}
}

// AddBlock adds a new block with the given content and options
func (pb *PromptBuilder) AddBlock(content string, opts ...BlockOption) *PromptBuilder {
	block := &Block{
		content: content,
		level:   1, // default heading level
	}

	for _, opt := range opts {
		opt(block)
	}

	pb.blocks = append(pb.blocks, block)
	return pb
}

// AddHeading adds a heading block
func (pb *PromptBuilder) AddHeading(text string, level int) *PromptBuilder {
	return pb.AddBlock("", WithTitle(text), WithLevel(level))
}

// AddH1 adds a level 1 heading
func (pb *PromptBuilder) AddH1(text string) *PromptBuilder {
	return pb.AddHeading(text, 1)
}

// AddH2 adds a level 2 heading
func (pb *PromptBuilder) AddH2(text string) *PromptBuilder {
	return pb.AddHeading(text, 2)
}

// AddH3 adds a level 3 heading
func (pb *PromptBuilder) AddH3(text string) *PromptBuilder {
	return pb.AddHeading(text, 3)
}

// AddH4 adds a level 4 heading
func (pb *PromptBuilder) AddH4(text string) *PromptBuilder {
	return pb.AddHeading(text, 4)
}

// AddH5 adds a level 5 heading
func (pb *PromptBuilder) AddH5(text string) *PromptBuilder {
	return pb.AddHeading(text, 5)
}

// AddH6 adds a level 6 heading
func (pb *PromptBuilder) AddH6(text string) *PromptBuilder {
	return pb.AddHeading(text, 6)
}

// AddParagraph adds a paragraph block
func (pb *PromptBuilder) AddParagraph(content string) *PromptBuilder {
	return pb.AddBlock(content)
}

// AddCodeBlock adds a code block with optional language
func (pb *PromptBuilder) AddCodeBlock(code, language string) *PromptBuilder {
	content := fmt.Sprintf("```%s\n%s\n```", language, code)
	return pb.AddBlock(content)
}

// AddList adds a list block
func (pb *PromptBuilder) AddList(items []string, ordered bool) *PromptBuilder {
	var listItems []string
	for i, item := range items {
		if ordered {
			listItems = append(listItems, fmt.Sprintf("%d. %s", i+1, item))
		} else {
			listItems = append(listItems, fmt.Sprintf("- %s", item))
		}
	}
	content := strings.Join(listItems, "\n")
	return pb.AddBlock(content)
}

// AddUnorderedList adds an unordered list
func (pb *PromptBuilder) AddUnorderedList(items []string) *PromptBuilder {
	return pb.AddList(items, false)
}

// AddOrderedList adds an ordered list
func (pb *PromptBuilder) AddOrderedList(items []string) *PromptBuilder {
	return pb.AddList(items, true)
}

// AddBlockquote adds a blockquote
func (pb *PromptBuilder) AddBlockquote(content string) *PromptBuilder {
	lines := strings.Split(content, "\n")
	var quotedLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			quotedLines = append(quotedLines, fmt.Sprintf("> %s", line))
		}
	}
	return pb.AddBlock(strings.Join(quotedLines, "\n"))
}

// AddSeparator adds a horizontal rule
func (pb *PromptBuilder) AddSeparator() *PromptBuilder {
	return pb.AddBlock("---")
}

// AddProtoBlock adds a block with marshaled proto.Message as YAML
func (pb *PromptBuilder) AddProtoBlock(msg proto.Message, opts ...BlockOption) *PromptBuilder {
	content := MarshalProtoToYAMLSafe(msg)
	return pb.AddBlock(content, opts...)
}

func (pb *PromptBuilder) AddProtosBlock(msgs ...proto.Message) *PromptBuilder {
	content := ""
	for _, msg := range msgs {
		content += MarshalProtoToYAMLSafe(msg) + "\n"
	}
	return pb.AddBlock(content)
}

// Build renders all blocks to markdown
func (pb *PromptBuilder) Build() string {
	var result []string

	for _, block := range pb.blocks {
		var blockContent []string

		// Add title if present
		if block.title != "" {
			hashes := strings.Repeat("#", block.level)
			blockContent = append(blockContent, fmt.Sprintf("%s %s", hashes, block.title))
			// Add a blank line after the heading if there is content
			if block.content != "" {
				blockContent = append(blockContent, "")
			}
		}

		// Add content if present
		if block.content != "" {
			blockContent = append(blockContent, block.content)
		}

		// Join block content
		if len(blockContent) > 0 {
			result = append(result, strings.Join(blockContent, "\n"))
		}
	}

	return strings.Join(result, "\n\n")
}

func (pb *PromptBuilder) BuildUserMessage() *ai.TextMessage {
	return ai.NewUserMessage(pb.Build())
}

// String implements the Stringer interface
func (pb *PromptBuilder) String() string {
	return pb.Build()
}

// Clear removes all blocks from the builder
func (pb *PromptBuilder) Clear() *PromptBuilder {
	pb.blocks = make([]*Block, 0)
	return pb
}

// BlockCount returns the number of blocks in the builder
func (pb *PromptBuilder) BlockCount() int {
	return len(pb.blocks)
}

// GetBlocks returns a copy of all blocks
func (pb *PromptBuilder) GetBlocks() []*Block {
	blocks := make([]*Block, len(pb.blocks))
	copy(blocks, pb.blocks)
	return blocks
}

func ConvertToProtoMessages[T proto.Message](msgs []T) []proto.Message {
	result := make([]proto.Message, len(msgs))
	for i, msg := range msgs {
		result[i] = msg
	}
	return result
}

// HELPERS

// marshalProtoToYAML converts a protobuf message to YAML by first marshaling to JSON
func MarshalProtoToYAML(msg proto.Message) ([]byte, error) {
	// First marshal to JSON with proper field names
	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Convert JSON to a map to preserve field names
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		return nil, err
	}

	// Marshal the map to YAML
	return yaml.Marshal(jsonMap)
}

// MarshalProtoToYAMLSafe safely marshals a protobuf message to YAML
// This function handles errors gracefully and returns a fallback string if marshaling fails
func MarshalProtoToYAMLSafe(msg proto.Message) string {
	if msg == nil {
		return "null"
	}

	// Try to marshal using the standard approach
	yamlBytes, err := MarshalProtoToYAML(msg)
	if err == nil {
		return string(yamlBytes)
	}

	// If that fails, try a simpler approach using protojson directly
	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("Failed to marshal proto: %v", err)
	}

	// Convert JSON to YAML manually
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		return fmt.Sprintf("Failed to parse JSON: %v", err)
	}

	yamlBytes, err = yaml.Marshal(jsonMap)
	if err != nil {
		return fmt.Sprintf("Failed to marshal to YAML: %v", err)
	}

	return string(yamlBytes)
}
