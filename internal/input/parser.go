package input

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// JobInfo contains parsed job posting information
type JobInfo struct {
	Company     string
	Title       string
	Description string
	URL         string
}

// ParseExperience reads and returns the experience.md content
func ParseExperience(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read experience file: %w", err)
	}

	content := string(data)
	if len(content) == 0 {
		return "", fmt.Errorf("experience file is empty")
	}

	return content, nil
}

// ParseJobDescription extracts company, role, and requirements from job description text
// Expects format like:
// Company: Acme Corp
// Role: Senior Full Stack Developer
// <rest is description>
func ParseJobDescription(jobText string) (*JobInfo, error) {
	if len(jobText) == 0 {
		return nil, fmt.Errorf("job description is empty")
	}

	jobInfo := &JobInfo{
		Description: jobText,
	}

	lines := strings.Split(jobText, "\n")

	// Try to extract Company and Role from first few lines
	for i, line := range lines {
		if i > 10 { // Only check first 10 lines
			break
		}

		line = strings.TrimSpace(line)

		// Check for "Company:" pattern
		if strings.HasPrefix(strings.ToLower(line), "company:") {
			jobInfo.Company = strings.TrimSpace(strings.TrimPrefix(line, "Company:"))
			jobInfo.Company = strings.TrimSpace(strings.TrimPrefix(jobInfo.Company, "company:"))
			continue
		}

		// Check for "Role:" or "Title:" or "Position:" pattern
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "role:") || strings.HasPrefix(lower, "title:") || strings.HasPrefix(lower, "position:") {
			jobInfo.Title = strings.TrimSpace(regexp.MustCompile(`^(role|title|position):\s*`, ).ReplaceAllString(line, ""))
			continue
		}

		// Check for "URL:" pattern
		if strings.HasPrefix(strings.ToLower(line), "url:") {
			jobInfo.URL = strings.TrimSpace(strings.TrimPrefix(line, "URL:"))
			jobInfo.URL = strings.TrimSpace(strings.TrimPrefix(jobInfo.URL, "url:"))
			continue
		}
	}

	// If Company or Title not found, use defaults
	if jobInfo.Company == "" {
		jobInfo.Company = extractCompanyFromDescription(jobText)
	}
	if jobInfo.Title == "" {
		jobInfo.Title = extractTitleFromDescription(jobText)
	}

	return jobInfo, nil
}

// extractCompanyFromDescription attempts to find company name in description text
func extractCompanyFromDescription(text string) string {
	// Look for common patterns like "at [Company]" or "join [Company]"
	patterns := []string{
		`(?i)at ([A-Z][A-Za-z0-9\s&]+?)(?:,|\.|$)`,
		`(?i)join ([A-Z][A-Za-z0-9\s&]+?)(?:,|\.|$)`,
		`(?i)([A-Z][A-Za-z0-9\s&]+?) is (?:looking|seeking|hiring)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			company := strings.TrimSpace(matches[1])
			if len(company) > 2 && len(company) < 50 {
				return company
			}
		}
	}

	return "Target Company"
}

// extractTitleFromDescription attempts to find job title in description text
func extractTitleFromDescription(text string) string {
	// Look for common patterns
	patterns := []string{
		`(?i)(?:position|role|title):\s*([A-Z][A-Za-z0-9\s/]+?)(?:,|\.|$)`,
		`(?i)(?:seeking|hiring|looking for)(?:\s+a?)?\s+([A-Z][A-Za-z0-9\s/]+?)(?:\s+to|\s+who|,|\.|$)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if len(title) > 5 && len(title) < 100 {
				return title
			}
		}
	}

	return "Software Engineer"
}

// CreateJobInfo creates a JobInfo struct from separate fields (for manual input)
func CreateJobInfo(company, title, description, url string) *JobInfo {
	// Set defaults if empty
	if company == "" {
		company = "Target Company"
	}
	if title == "" {
		title = "Software Engineer"
	}

	return &JobInfo{
		Company:     company,
		Title:       title,
		Description: description,
		URL:         url,
	}
}
