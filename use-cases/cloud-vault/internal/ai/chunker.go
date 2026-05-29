package ai

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode/utf8"
)

// chunkDocument splits Markdown content into heading-bounded chunks.
// Each chunk aims for ~800 tokens. Code fences are never split mid-block.
func chunkDocument(documentID, content string) []*DocumentChunk {
	lines := strings.Split(content, "\n")
	type segment struct {
		headingPath string
		lines       []string
	}

	var segments []segment
	current := segment{}
	var headingStack []string
	inFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks.
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
		}

		if !inFence && strings.HasPrefix(trimmed, "#") {
			// Flush previous segment.
			if len(current.lines) > 0 {
				segments = append(segments, current)
			}
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			heading := strings.TrimSpace(trimmed[level:])
			// Maintain heading stack for path.
			for len(headingStack) >= level {
				headingStack = headingStack[:len(headingStack)-1]
			}
			headingStack = append(headingStack, heading)
			current = segment{
				headingPath: strings.Join(headingStack, " > "),
				lines:       []string{line},
			}
			continue
		}

		current.lines = append(current.lines, line)
	}
	if len(current.lines) > 0 {
		segments = append(segments, current)
	}

	// Merge or split segments to target ~800 tokens each.
	const targetTokens = 800
	const maxTokens = 1500

	var chunks []*DocumentChunk
	chunkIdx := 0
	accLines := []string{}
	accPath := ""

	flush := func() {
		if len(accLines) == 0 {
			return
		}
		text := strings.Join(accLines, "\n")
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		tokens := estimateTokens(text)
		hp := accPath
		chunks = append(chunks, &DocumentChunk{
			DocumentID:  documentID,
			ChunkIndex:  chunkIdx,
			HeadingPath: nilIfEmpty(hp),
			Content:     text,
			ContentHash: hashContent(text),
			TokenCount:  tokens,
		})
		chunkIdx++
		accLines = []string{}
		accPath = ""
	}

	for _, seg := range segments {
		text := strings.Join(seg.lines, "\n")
		tok := estimateTokens(text)

		if estimateTokens(strings.Join(accLines, "\n"))+tok > maxTokens {
			flush()
		}
		if accPath == "" {
			accPath = seg.headingPath
		}
		accLines = append(accLines, seg.lines...)

		if estimateTokens(strings.Join(accLines, "\n")) >= targetTokens {
			flush()
		}
	}
	flush()

	if len(chunks) == 0 && utf8.RuneCountInString(content) > 0 {
		chunks = append(chunks, &DocumentChunk{
			DocumentID:  documentID,
			ChunkIndex:  0,
			Content:     content,
			ContentHash: hashContent(content),
			TokenCount:  estimateTokens(content),
		})
	}

	return chunks
}

func hashContent(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
