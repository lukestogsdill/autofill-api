package llm

// DynamicResumeContent contains only the LLM-generated fields (marked with # in template)
type DynamicResumeContent struct {
	Title                  string          `json:"title"`
	Summary                string          `json:"summary"`
	Skills                 []SkillCategory `json:"skills"`
	ExperienceAchievements [][]Achievement `json:"experience_achievements"`
	ProjectsAchievements   [][]Achievement `json:"projects_achievements"`
}

// SkillCategory represents a skill category with items
type SkillCategory struct {
	Category string `json:"category"`
	Items    string `json:"items"`
	Overflow bool   `json:"overflow,omitempty"`
}

// Achievement represents a single bullet point
type Achievement struct {
	Text     string `json:"text"`
	Overflow bool   `json:"overflow,omitempty"`
}

// ResumeContentSchema defines the structured output schema for LLM-generated resume content
// Only dynamic fields - static data comes from resume.json
var ResumeContentSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title": map[string]any{
			"type":        "string",
			"description": "Professional title tailored to the job",
		},
		"summary": map[string]any{
			"type":        "string",
			"description": "Two sentence summary emphasizing fit for this role",
		},
		"skills": map[string]any{
			"type":        "array",
			"description": "4 categories maximum, 6-8 skills per category",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{
						"type":        "string",
						"description": "Keep categories short (e.g. Frontend, Backend, Databases, DevOps)",
					},
					"items": map[string]any{
						"type":        "string",
						"description": "Comma-separated list of skills",
					},
					"overflow": map[string]any{"type": "boolean"},
				},
				"required": []string{"category", "items"},
			},
		},
		"experience_achievements": map[string]any{
			"type":        "array",
			"description": "Array of achievement arrays - one per experience entry (2 experiences total)",
			"items": map[string]any{
				"type":        "array",
				"description": "3-4 achievements for this experience",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "Achievement with metrics and impact",
						},
						"overflow": map[string]any{"type": "boolean"},
					},
					"required": []string{"text"},
				},
			},
		},
		"projects_achievements": map[string]any{
			"type":        "array",
			"description": "Array of achievement arrays - one per project (2 projects total)",
			"items": map[string]any{
				"type":        "array",
				"description": "3 achievements for this project",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "Key feature or impact",
						},
						"overflow": map[string]any{"type": "boolean"},
					},
					"required": []string{"text"},
				},
			},
		},
	},
	"required": []string{"title", "summary", "skills", "experience_achievements", "projects_achievements"},
}

// UniversalInfoSchema defines the schema for common application questions
var UniversalInfoSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"why_work_here": map[string]any{
			"type":        "string",
			"description": "Why you want to work at this company (2-3 sentences)",
		},
		"why_interested": map[string]any{
			"type":        "string",
			"description": "Why you're interested in this role (2-3 sentences)",
		},
		"salary_expectations": map[string]any{
			"type":        "string",
			"description": "Salary expectations",
		},
		"start_date": map[string]any{
			"type":        "string",
			"description": "When you can start",
		},
		"leaving_reason": map[string]any{
			"type":        "string",
			"description": "Why you're leaving current position (2-3 sentences)",
		},
		"strengths": map[string]any{
			"type":        "string",
			"description": "Key strengths (2-3 sentences)",
		},
		"weaknesses": map[string]any{
			"type":        "string",
			"description": "Area you're working to improve (2-3 sentences)",
		},
	},
	"required": []string{"why_work_here", "why_interested", "salary_expectations", "start_date", "leaving_reason", "strengths", "weaknesses"},
}
