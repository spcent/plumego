package ai

import (
	"fmt"
	"strings"
	"time"
)

// WriteSummaryMarkdown converts a SummaryOutput into a saveable Markdown document.
func WriteSummaryMarkdown(title, providerName, model string, out *SummaryOutput, docIDs []string) string {
	var sb strings.Builder
	sb.WriteString("# " + title + "\n\n")
	sb.WriteString(fmt.Sprintf("> AI-generated summary · provider: %s · model: %s · %s\n\n",
		providerName, model, time.Now().UTC().Format("2006-01-02")))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(out.Summary + "\n\n")

	if len(out.KeyPoints) > 0 {
		sb.WriteString("## Key Points\n\n")
		for _, p := range out.KeyPoints {
			sb.WriteString("- " + p + "\n")
		}
		sb.WriteString("\n")
	}

	if len(out.Actions) > 0 {
		sb.WriteString("## Action Items\n\n")
		for _, a := range out.Actions {
			sb.WriteString("- " + a + "\n")
		}
		sb.WriteString("\n")
	}

	if len(out.CodeRefs) > 0 {
		sb.WriteString("## Code References\n\n")
		for _, c := range out.CodeRefs {
			sb.WriteString("- " + c + "\n")
		}
		sb.WriteString("\n")
	}

	if len(docIDs) > 0 {
		sb.WriteString("## Source Documents\n\n")
		for _, id := range docIDs {
			sb.WriteString(fmt.Sprintf("- `%s`\n", id))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// WriteQAMarkdown converts a QAOutput into a saveable Markdown document.
func WriteQAMarkdown(question, providerName, model string, out *QAOutput) string {
	var sb strings.Builder
	sb.WriteString("# Q&A: " + truncateTitle(question, 60) + "\n\n")
	sb.WriteString(fmt.Sprintf("> AI-generated answer · provider: %s · model: %s · %s\n\n",
		providerName, model, time.Now().UTC().Format("2006-01-02")))

	sb.WriteString("## Question\n\n")
	sb.WriteString(question + "\n\n")

	sb.WriteString("## Answer\n\n")
	sb.WriteString(out.Answer + "\n\n")

	if len(out.Citations) > 0 {
		sb.WriteString("## Citations\n\n")
		for _, c := range out.Citations {
			sb.WriteString(fmt.Sprintf("**%s** (`%s`)\n\n> %s\n\n", c.DocumentTitle, c.DocumentID, c.Excerpt))
		}
	}

	if len(out.DocumentIDs) > 0 {
		sb.WriteString("## Source Documents\n\n")
		for _, id := range out.DocumentIDs {
			sb.WriteString(fmt.Sprintf("- `%s`\n", id))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// WritePromptMarkdown converts a PromptExtractOutput into a saveable Markdown document.
func WritePromptMarkdown(out *PromptExtractOutput, sourceDocID, providerName, model string) string {
	var sb strings.Builder
	sb.WriteString("# " + out.Title + "\n\n")
	sb.WriteString(fmt.Sprintf("> Extracted prompt · provider: %s · model: %s · %s\n\n",
		providerName, model, time.Now().UTC().Format("2006-01-02")))

	if out.Scenario != "" {
		sb.WriteString(fmt.Sprintf("**Scenario:** %s\n\n", out.Scenario))
	}
	if out.ModelHint != "" {
		sb.WriteString(fmt.Sprintf("**Recommended model:** %s\n\n", out.ModelHint))
	}
	if out.QualityScore > 0 {
		sb.WriteString(fmt.Sprintf("**Quality score:** %.2f\n\n", out.QualityScore))
	}

	sb.WriteString("## Prompt\n\n```\n")
	sb.WriteString(out.Content)
	sb.WriteString("\n```\n\n")

	if sourceDocID != "" {
		sb.WriteString("## Source\n\n")
		sb.WriteString(fmt.Sprintf("- Source document: `%s`\n\n", sourceDocID))
	}

	return sb.String()
}

func truncateTitle(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
