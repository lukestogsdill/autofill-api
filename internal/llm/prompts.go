package llm

import (
	"fmt"
	"strings"
)

// JobInfo contains job posting details
type JobInfo struct {
	Company     string
	Title       string
	Description string
	URL         string
}

// ResumeGenerationPrompt creates the prompt for tailoring a resume to a job
func ResumeGenerationPrompt(experienceMd string, job JobInfo) string {
	return fmt.Sprintf(`You are a professional resume writer tailoring a resume for a specific job application.

Source experience (in Markdown format):
%s

Target job:
- Company: %s
- Role: %s
- Requirements: %s

Generate ONLY the dynamic resume content customized for this job:

EXPERIENCE (2 positions):
1. Lobby Media | Web Design - Co-Founder (June 2024 - Current)
   Tech: Svelte5, Next.js, TypeScript, SveltiaCS, Cloudflare

2. Freelance Software Consultant - Contract Developer (December 2023 - June 2024)
   Tech: Next.js, Golang, PostgreSQL, MongoDB, Docker, Hetzner Cloud

PROJECTS (2 projects):
1. Sibyl - Daily Morality Quiz
   Tech: React, TypeScript, Convex, TanStack Query, oRPC

2. AI VTuber Content Creator
   Tech: React, Python, C++, Live2D Cubism SDK, WhisperX, OpenAI API

Generate:
1. title: Professional title matching the job, no sentences please, title only
2. summary: Exactly 2 sentences emphasizing fit for this role
3. skills: 4 categories max (Frontend, Backend, Databases, DevOps), 4-5 skills each
4. experience_achievements: Array with 2 sub-arrays, 3-4 achievements each for the experiences above
5. projects_achievements: Array with 2 sub-arrays, 3 achievements each for the projects above

Prioritize impact, metrics, and relevance to the job requirements.`,
		experienceMd, job.Company, job.Title, truncate(job.Description, 2000))
}

// CoverLetterPrompt creates the prompt for generating a cover letter
func CoverLetterPrompt(resumeSummary string, keyAchievements []string, job JobInfo) string {
	achievements := strings.Join(keyAchievements, "\n- ")

	return fmt.Sprintf(`Write a professional cover letter for this job application.

Job: %s at %s
Requirements: %s

Your background:
Summary: %s

Key achievements:
- %s

Cover letter requirements:
- 3-4 paragraphs maximum
- Address specific job requirements from the description
- Highlight 2-3 most relevant achievements
- Show genuine enthusiasm for the company and role
- Professional but personable tone
- NO generic templates - make it specific to THIS job

Output the cover letter text only (no "Dear Hiring Manager" salutation needed).`,
		job.Title, job.Company, truncate(job.Description, 1500),
		resumeSummary, achievements)
}

// UniversalInfoPrompt creates the prompt for common application questions
func UniversalInfoPrompt(resumeSummary string, job JobInfo) string {
	return fmt.Sprintf(`Generate concise answers for common job application questions based on this profile.

Job: %s at %s
Your background: %s

Generate answers for these questions (2-3 sentences each):
1. why_work_here: Why do you want to work at this company?
2. why_interested: Why are you interested in this role?
3. salary_expectations: What are your salary expectations?
4. start_date: When can you start?
5. leaving_reason: Why are you leaving your current position?
6. strengths: What are your key strengths?
7. weaknesses: What is an area you're working to improve?

Return as JSON object:
{
  "why_work_here": "answer",
  "why_interested": "answer",
  "salary_expectations": "answer",
  "start_date": "answer",
  "leaving_reason": "answer",
  "strengths": "answer",
  "weaknesses": "answer"
}

CRITICAL: Output ONLY valid JSON. No explanations.`,
		job.Title, job.Company, resumeSummary)
}

// FieldFillingPrompt creates the prompt for filling unknown form fields
func FieldFillingPrompt(fieldLabel string, fieldType string, options []string, context map[string]string) string {
	optionsText := ""
	if len(options) > 0 {
		optionsText = fmt.Sprintf("\nAvailable options: %s", strings.Join(options, ", "))
	}

	return fmt.Sprintf(`You are filling out a job application form.

Field label: %s
Field type: %s%s

Context:
- Job: %s at %s
- Your summary: %s
- Your skills: %s

Provide a concise, appropriate answer for this field.
If this is a select/radio field, choose the best option from the available options.
Output ONLY the answer text (no explanation).`,
		fieldLabel, fieldType, optionsText,
		context["job_title"], context["company"],
		context["summary"], context["skills"])
}

// ColdMessagePrompt creates the prompt for generating a cold outreach message
func ColdMessagePrompt(experienceSummary string, keyAchievements []string, job JobInfo) string {
	achievements := strings.Join(keyAchievements, "\n- ")

	return fmt.Sprintf(`Write a cold outreach message for this job opportunity.

Job: %s at %s
Job description: %s

Your background:
%s

Key achievements:
- %s

Message requirements:
- Very concise: 3-4 short paragraphs maximum
- Professional but friendly tone
- Start with a compelling hook about why you're reaching out
- Highlight 1-2 most relevant achievements that match the role
- Express genuine interest in the company/role
- Include a clear call-to-action (request for a brief chat/interview)
- NO generic templates - make it specific and personal
- Keep it conversational, not salesy

Output the message text only (no subject line needed).`,
		job.Title, job.Company, truncate(job.Description, 1500),
		experienceSummary, achievements)
}

// truncate limits string length for token management
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
