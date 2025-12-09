package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"autofill-api/internal/llm"
)

// GenerateResumeJSON uses LLM to tailor resume content to a specific job
func GenerateResumeJSON(experienceMd string, jobInfo llm.JobInfo, llmClient *llm.Client) (*Resume, error) {
	log.Printf("Generating resume for %s at %s...", jobInfo.Title, jobInfo.Company)

	// Create prompt for dynamic content only
	prompt := llm.ResumeGenerationPrompt(experienceMd, jobInfo)

	// Generate dynamic resume content using LLM with structured output
	var dynamicContent llm.DynamicResumeContent
	err := llmClient.GenerateJSON(prompt, &dynamicContent, llm.ResumeContentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate resume content: %w", err)
	}

	// Build full resume by merging dynamic content with static data
	resume := buildResumeFromDynamic(&dynamicContent)

	// Validate resume has required fields
	if err := validateResume(resume); err != nil {
		return nil, fmt.Errorf("generated resume is invalid: %w", err)
	}

	log.Printf("Resume generated successfully (%d experiences, %d projects, %d skills)",
		len(resume.Experience), len(resume.Projects), len(resume.Skills))

	return resume, nil
}

// buildResumeFromDynamic loads resume.json and replaces dynamic fields with LLM output
func buildResumeFromDynamic(dynamic *llm.DynamicResumeContent) *Resume {
	// Load static resume template
	resumeData, err := os.ReadFile("resume.json")
	if err != nil {
		log.Printf("Warning: failed to load resume.json: %v", err)
		// Return a minimal resume if file doesn't exist
		return &Resume{
			Name:       "Luke Stogsdill",
			Title:      dynamic.Title,
			Summary:    dynamic.Summary,
			Skills:     convertSkills(dynamic.Skills),
			Experience: []Experience{},
			Projects:   []Project{},
			Education:  []Education{},
		}
	}

	var resume Resume
	if err := json.Unmarshal(resumeData, &resume); err != nil {
		log.Printf("Warning: failed to parse resume.json: %v", err)
		return &Resume{}
	}

	// Replace dynamic fields with LLM-generated content
	resume.Title = dynamic.Title
	resume.Summary = dynamic.Summary
	resume.Skills = convertSkills(dynamic.Skills)

	// Replace achievements for experiences
	for i, achievements := range dynamic.ExperienceAchievements {
		if i < len(resume.Experience) {
			resume.Experience[i].Achievements = convertAchievements(achievements)
		}
	}

	// Replace achievements for projects
	for i, achievements := range dynamic.ProjectsAchievements {
		if i < len(resume.Projects) {
			resume.Projects[i].Achievements = convertAchievements(achievements)
		}
	}

	return &resume
}

// convertSkills converts llm.SkillCategory to generator.Skill
func convertSkills(source []llm.SkillCategory) []Skill {
	result := make([]Skill, len(source))
	for i, sk := range source {
		result[i] = Skill{
			Category: sk.Category,
			Items:    sk.Items,
			Overflow: sk.Overflow,
		}
	}
	return result
}

// convertAchievements converts llm.Achievement to generator.Achievement
func convertAchievements(source []llm.Achievement) []Achievement {
	result := make([]Achievement, len(source))
	for i, a := range source {
		result[i] = Achievement{
			Text:     a.Text,
			Overflow: a.Overflow,
		}
	}
	return result
}

// GenerateCoverLetter creates a tailored cover letter
func GenerateCoverLetter(resume *Resume, jobInfo llm.JobInfo, llmClient *llm.Client) (string, error) {
	log.Printf("Generating cover letter for %s at %s...", jobInfo.Title, jobInfo.Company)

	// Extract key achievements from resume
	keyAchievements := extractKeyAchievements(resume)

	// Create prompt
	prompt := llm.CoverLetterPrompt(resume.Summary, keyAchievements, jobInfo)

	// Generate cover letter
	coverLetter, err := llmClient.GenerateText(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate cover letter: %w", err)
	}

	log.Printf("Cover letter generated (%d chars)", len(coverLetter))

	return coverLetter, nil
}

// ApplicationInfo contains answers to common application questions
type ApplicationInfo struct {
	WhyWorkHere         string `json:"why_work_here"`
	WhyInterested       string `json:"why_interested"`
	SalaryExpectations  string `json:"salary_expectations"`
	StartDate           string `json:"start_date"`
	LeavingReason       string `json:"leaving_reason"`
	Strengths           string `json:"strengths"`
	Weaknesses          string `json:"weaknesses"`
}

// GenerateUniversalInfo creates answers for common application questions
func GenerateUniversalInfo(resume *Resume, jobInfo llm.JobInfo, llmClient *llm.Client) (*ApplicationInfo, error) {
	log.Printf("Generating universal application info...")

	// Create prompt
	prompt := llm.UniversalInfoPrompt(resume.Summary, jobInfo)

	// Generate answers with structured output
	var appInfo ApplicationInfo
	err := llmClient.GenerateJSON(prompt, &appInfo, llm.UniversalInfoSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate application info: %w", err)
	}

	log.Printf("Universal info generated (7 questions answered)")

	return &appInfo, nil
}

// validateResume checks that required fields are present
func validateResume(resume *Resume) error {
	if resume.Name == "" {
		return fmt.Errorf("missing name")
	}
	if resume.Title == "" {
		return fmt.Errorf("missing title")
	}
	if resume.Summary == "" {
		return fmt.Errorf("missing summary")
	}
	if len(resume.Skills) == 0 {
		return fmt.Errorf("missing skills")
	}
	if len(resume.Experience) == 0 {
		return fmt.Errorf("missing experience")
	}

	return nil
}

// extractKeyAchievements gets the most impressive achievements from resume
func extractKeyAchievements(resume *Resume) []string {
	achievements := []string{}

	// Get first 2 achievements from each experience
	for _, exp := range resume.Experience {
		for i, ach := range exp.Achievements {
			if i >= 2 { // Max 2 per experience
				break
			}
			achievements = append(achievements, ach.Text)
		}
	}

	// Get first achievement from each project
	for _, proj := range resume.Projects {
		if len(proj.Achievements) > 0 {
			achievements = append(achievements, proj.Achievements[0].Text)
		}
	}

	// Limit to 5 total achievements
	if len(achievements) > 5 {
		achievements = achievements[:5]
	}

	return achievements
}
