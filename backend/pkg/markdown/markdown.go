package markdown

import (
	"fmt"
	"strings"
	"time"

	"github.com/alex/reading-companion/internal/models"
)

// GenerateMarkdown generates a markdown document from session data
func GenerateMarkdown(session *models.Session, highlights []*models.Highlight, interactions map[string]*models.Interaction) string {
	var sb strings.Builder

	// Add header
	sb.WriteString(fmt.Sprintf("# %s\n", session.Name))
	sb.WriteString(fmt.Sprintf("**Дата разбора:** %s\n\n", time.Now().Format("02.01.2006")))
	sb.WriteString("---\n\n")

	// Add each highlight with its interaction
	for _, highlight := range highlights {
		// Add the highlight as a quote
		sb.WriteString(fmt.Sprintf("> %s\n\n", highlight.Text))

		// Add the interaction if it exists
		if interaction, exists := interactions[highlight.ID.String()]; exists {
			// Add the question
			sb.WriteString(fmt.Sprintf("**_Вопрос ассистента: %s_**\n\n", interaction.Question))

			// Add the answer if it exists
			if interaction.Answer.Valid {
				sb.WriteString(fmt.Sprintf("%s\n\n", interaction.Answer.String))
			}
		}

		// Add a separator
		sb.WriteString("---\n\n")
	}

	return sb.String()
}
