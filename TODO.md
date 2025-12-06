# Autofill API - Complete Rewrite

## Architecture Overview

**Goal:** Generate tailored resumes, cover letters, and autofill application forms using LLM + constant fallbacks.

**Workflow:**
```
1. Input: experience.md + job description/company info (manual or scraped)
2. LLM generates role-specific resume content (formatted for maroto)
3. Generate PDF resume using existing Go maroto templating (with 1-page constraint)
4. Generate cover letter + universal application info
5. Navigate to job application form
6. Bookmarklet scrapes form fields
7. API fills fields:
   - Check constants (first_name, email, etc.)
   - Fallback to LLM with resume generation context
8. Update Google Doc with application record
```

**Data Flow:**
```
experience.md (source of truth)
    ↓
LLM tailors content for job → resume.json
    ↓
main.go generates PDF → resume.pdf (1 page max)
    ↓
LLM generates cover letter + universal info
    ↓
Form filling (constants → LLM fallback)
    ↓
Google Docs API (tracking)
```

**Storage:**
- **No database** - everything derived from experience.md + LLM
- **Constants file** (JSON): first_name, last_name, email, phone, etc.
- **Google Docs** for application history tracking

---

## Phase 1: Resume Generation Pipeline

**Approach:**
1. **Single LLM call** with strict content constraints (no regeneration loops)
2. **Proportional scaling** if PDF exceeds 1 page (scales fonts/spacing, not content)
3. **Manual job description input** (copy/paste from LinkedIn/Indeed)

### Input Processing
- [ ] Create `internal/input/parser.go`:
  - `ParseExperience(experienceMd string) (Experience, error)` - parse experience.md
  - `ParseJobDescription(jobDesc string) (JobInfo, error)` - extract company, role, requirements from manual paste
    - Job description will be manual paste from LinkedIn/Indeed (no web scraping needed)
!# well put this in like job-description.txt 

### LLM Content Generation
- [ ] Install Gemini SDK: `go get google.golang.org/api/ai`
- [ ] Set up API key in `.env`: `GEMINI_API_KEY=your_key`
- [ ] Create `internal/generator/resume_generator.go`:
  - `GenerateResumeJSON(experience Experience, jobInfo JobInfo) (Resume, error)`:
    - Takes experience.md content + job description
    - Uses LLM to tailor achievements, projects, skills for this specific role
    - Returns Resume struct (compatible with existing maroto code)
    - **Single LLM call with strict constraints** - no regeneration loops
  - Prompt template with strict content limits:
    ```
    "You are tailoring a resume for a specific job application.

    Source experience: {experience.md content}

    Target job:
    - Company: {company}
    - Role: {title}
    - Requirements: {job_description}

    Generate a resume.json that follows these STRICT CONSTRAINTS:
    1. Work Experience: Include ONLY the 1-2 most relevant positions
    2. Each position: Maximum 3 bullet points (choose most impactful)
    3. Projects: Maximum 2 projects total
    4. Each project: Maximum 3 bullet points
    5. Skills: 4 categories maximum, 6-8 skills per category
    6. Summary: Exactly 2 sentences
    7. Total estimated content: ~250-300 words
    8. Use exact format from provided schema
    9. Prioritize impact over completeness

    Mark longer text items with 'overflow: true' for proper spacing.

    Output JSON only."
    ```

### Resume PDF Generation (Existing System)
- [ ] Integrate existing maroto code from resume-builder
- [ ] Create `internal/generator/pdf_generator.go`:
  - Move `getMaroto()` function from main.go
  - Move `getPrimaryColor()` helper
  - `GeneratePDF(resume Resume, outputPath string) error`
  - Add configurable scaling parameters (fontSizes, rowHeights, margins)
- [ ] Add 1-page constraint with proportional scaling:
  - Generate PDF with normal sizes (attempt 1)
  - Check if page count > 1
  - If overflow: Apply scaling algorithm (max 3 attempts)
    ```go
    scaleFactor := 0.95  // Start with 5% reduction
    for attempt := 1; attempt <= 3; attempt++ {
        if pageCount <= 1 { break }

        // Scale all dimensions proportionally
        config.FontSizes = scaleMap(config.FontSizes, scaleFactor)
        config.RowHeights = scaleMap(config.RowHeights, scaleFactor)
        config.Margins = scaleMargins(config.Margins, scaleFactor)

        // Regenerate PDF with scaled dimensions
        GeneratePDF(resume, config)
        scaleFactor -= 0.05  // Reduce by additional 5% if needed
    }
    ```
  - Keep aspect ratios and readability (minimum font size: 8pt)
- [ ] Copy font files and icons to project:
  - `fonts/DejaVuSans*.ttf`
  - `icons-png/*.png`

### Cover Letter Generation
- [ ] Create `internal/generator/cover_letter_generator.go`:
  - `GenerateCoverLetter(resume Resume, jobInfo JobInfo) (string, error)`:
    - Uses LLM to write tailored cover letter
    - 3-4 paragraphs max
    - References specific job requirements
    - Highlights relevant achievements from resume
  - Prompt template:
    ```
    "Write a professional cover letter for this application.

    Job: {title} at {company}
    Requirements: {job_description}

    Your background (from resume):
    {resume summary + key achievements}

    Cover letter should:
    - Be 3-4 paragraphs
    - Address specific job requirements
    - Highlight 2-3 most relevant achievements
    - Show enthusiasm for company/role

    Output cover letter text only."
    ```

### Universal Application Info Generation
- [ ] Create `internal/generator/application_info.go`:
  - `GenerateUniversalInfo(resume Resume, jobInfo JobInfo) (ApplicationInfo, error)`:
    - Generates answers for common application questions:
      - "Why do you want to work here?"
      - "Why are you interested in this role?"
      - "What are your salary expectations?"
      - "When can you start?"
      - "Why are you leaving your current job?"
      - "What are your strengths/weaknesses?"
    - Returns structured data for later use in form filling
  - Store in memory during session (no persistence needed)

---

## Phase 2: Form Autofill System

### Constants Management
- [ ] Create `constants.json`:
  ```json
  {
    "first_name": "Luke",
    "last_name": "Stogsdill",
    "email": "lukestogsdill@gmail.com",
    "phone": "(832) 392-2613",
    "location": "Houston, TX 77064",
    "city": "Houston",
    "state": "TX",
    "zip": "77064",
    "country": "United States",
    "linkedin_url": "https://linkedin.com/in/luke-stogsdill",
    "github_url": "https://github.com/lukestogsdill",
    "website_url": "https://lustogs.com",
    "years_experience": "2",
    "willing_to_relocate": "yes",
    "require_sponsorship": "no",
    "authorized_to_work": "yes",
    "over_18": "yes"
  }
  ```
- [ ] Create `internal/constants/loader.go`:
  - `LoadConstants() (map[string]string, error)` - load from constants.json
  - `GetConstant(key string) (string, bool)` - retrieve constant value

### Field Matching Logic
- [ ] Create `internal/matcher/matcher.go`:
  - `MatchField(label string, constants map[string]string) (value string, source string, found bool)`:
    1. Normalize label (lowercase, trim, remove special chars)
    2. Check if label matches any constant key (fuzzy match)
    3. Return (value, "constant", true) if found
    4. Return ("", "", false) if not found
  - Fuzzy matching rules:
    - "first name" / "fname" / "given name" → first_name
    - "email address" / "e-mail" → email
    - "phone number" / "mobile" / "telephone" → phone

### LLM Fallback for Unknown Fields
- [ ] Create `internal/matcher/llm_matcher.go`:
  - `FillFieldWithLLM(label string, fieldType string, context FieldContext) (string, error)`:
    - Takes field label, type (text/select/radio), and context (job info, resume, cover letter)
    - Uses LLM to generate appropriate answer
    - Context includes: resume content, cover letter, universal application info
  - Prompt template:
    ```
    "You are filling out a job application form.

    Field label: {label}
    Field type: {type}
    Available options (if select/radio): {options}

    Context:
    - Job: {job_title} at {company}
    - Your resume summary: {resume.summary}
    - Your skills: {resume.skills}
    - Cover letter: {cover_letter}
    - Pre-generated answers: {universal_info}

    Provide a concise, appropriate answer for this field.
    Output answer only (no explanation)."
    ```

### Enhanced Bookmarklet Field Collection
- [ ] Update `script.js` to collect:
  - All input types: text, email, tel, url, date, number
  - Textareas
  - Select dropdowns (include all options)
  - Radio buttons (group by name, include all options)
  - Checkboxes
  - Field context:
    - Associated label text
    - Placeholder text
    - Section headings (h2/h3 above field)
    - Fieldset/legend text
    - aria-label
    - Help text
- [ ] Extract job context from page:
  - Job title (page title, h1, common selectors)
  - Company name (domain, page content, meta tags)
  - Job description (try common selectors: .description, .job-details, etc.)
- [ ] Send enhanced structure to API:
  ```javascript
  {
    fields: [
      {
        name: "fname",
        label: "First Name",
        type: "text",
        required: true,
        placeholder: "Enter your first name",
        context: "Personal Information"
      },
      {
        name: "cover_letter",
        label: "Cover Letter",
        type: "textarea",
        maxlength: 5000,
        required: false,
        context: "Application Materials"
      },
      {
        name: "years_exp",
        label: "Years of Experience",
        type: "select",
        options: ["0-1", "1-3", "3-5", "5-7", "7+"],
        required: true
      }
    ],
    job_context: {
      title: "Senior Full Stack Developer",
      company: "Acme Corp",
      description: "We are looking for..."
    }
  }
  ```

### API Fill Handler
- [ ] Create `POST /api/fill` endpoint:
  - Parse incoming request (fields + job_context)
  - If new job_context, trigger resume generation:
    1. Load experience.md
    2. Generate tailored resume.json with LLM
    3. Generate PDF
    4. Generate cover letter
    5. Generate universal application info
    6. Store in memory for this session
  - For each field:
    1. Try constant match first (MatchField)
    2. If no match, use LLM fallback (FillFieldWithLLM)
    3. Handle field type (text/select/radio/checkbox)
    4. Return appropriate value
  - Return response:
    ```json
    {
      "fields": {
        "fname": "Luke",
        "email": "lukestogsdill@gmail.com",
        "cover_letter": "Dear Hiring Manager...",
        "years_exp": "1-3",
        "why_this_company": "I'm excited about Acme Corp because..."
      },
      "metadata": {
        "constant_matches": 8,
        "llm_generated": 3,
        "resume_generated": true
      }
    }
    ```
- [ ] Handle select/radio fields:
  - If options provided, LLM picks best match from options
  - Example: years_exp options ["0-1", "1-3", "3-5"] → LLM chooses "1-3"
- [ ] Handle edge cases:
  - Multi-value fields (references: name1, email1, name2, email2)
  - Date formatting
  - Phone number formatting
  - Address fields (parse location constant)

---

## Phase 3: Application Tracking (Google Docs)

### Google Docs Integration
- [ ] Install Google Docs API library: `go get google.golang.org/api/docs/v1`
- [ ] Set up Google Cloud project:
  - Enable Google Docs API
  - Create service account
  - Download credentials JSON
  - Share target Google Doc with service account email
- [ ] Create `.env` entry: `GOOGLE_DOC_ID=your_doc_id`
- [ ] Create `internal/tracking/google_docs.go`:
  - `AppendApplication(application Application) error`:
    - Appends new entry to Google Doc
    - Format:
      ```
      ---
      Date: 2025-12-05
      Company: Acme Corp
      Role: Senior Full Stack Developer
      Resume Generated: ✓
      Cover Letter: [link to generated file]
      Status: Applied

      Job Description:
      {description}

      Fields Filled:
      - first_name: Luke (constant)
      - email: lukestogsdill@gmail.com (constant)
      - cover_letter: [LLM generated]
      - why_this_company: [LLM generated]
      ---
      ```
  - Uses Google Docs API to append formatted text
  - Include metadata: what was filled, source (constant vs LLM)

### Tracking Trigger
- [ ] After successful form fill, call `AppendApplication()`:
  - Include: company, role, job description, filled fields, sources
  - Include links to generated resume PDF / cover letter
  - Mark timestamp

---

## Phase 4: File Management

### Generated Files Organization
- [ ] Create directory structure:
  ```
  generated/
    ├── 2025-12-05_acme-corp_senior-fullstack/
    │   ├── resume.json
    │   ├── resume.pdf
    │   ├── cover_letter.txt
    │   └── application_info.json
    └── 2025-12-04_startup-xyz_frontend/
        ├── resume.json
        ├── resume.pdf
        ├── cover_letter.txt
        └── application_info.json
  ```
- [ ] Create `internal/storage/files.go`:
  - `CreateApplicationFolder(company, role, date) (path string, error)`
  - `SaveResume(path, content) error`
  - `SaveCoverLetter(path, content) error`
  - `SaveApplicationInfo(path, info) error`
  - `GetLatestApplication() (path string, error)` - for reusing recent generation

### Session Management
- [ ] Keep generated content in memory during API session:
  - Resume object
  - Cover letter text
  - Universal application info
  - Avoid regenerating for same job (use company+role as key)
- [ ] Add TTL (1 hour) - refresh if stale

---

## Phase 5: API Endpoints

### Core Endpoints
- [ ] `POST /api/generate` - Generate resume/cover letter for job
  - Input: job description, company name, role title
  - Output: resume.pdf path, cover_letter.txt path
  - Saves to generated/ folder
- [ ] `POST /api/fill` - Fill form fields (covered in Phase 2)
  - Calls /api/generate internally if new job
- [ ] `GET /api/constants` - Get all constants
- [ ] `POST /api/constants` - Update constants
- [ ] `GET /script.js` - Serve bookmarklet (existing)

### Resume Generation Endpoint Details
- [ ] Request format:
  ```json
  {
    "job_description": "We are looking for...",
    "company": "Acme Corp",
    "role": "Senior Full Stack Developer",
    "job_url": "https://acme.com/careers/123" // optional
  }
  ```
- [ ] Response format:
  ```json
  {
    "resume_pdf": "generated/2025-12-05_acme-corp_senior-fullstack/resume.pdf",
    "cover_letter": "generated/2025-12-05_acme-corp_senior-fullstack/cover_letter.txt",
    "session_id": "abc123",
    "page_count": 1,
    "success": true
  }
  ```

---

## Phase 6: 1-Page Resume Constraint (Proportional Scaling)

### Scaling Configuration
- [ ] Create `internal/generator/config.go`:
  - Define `PDFConfig` struct with configurable dimensions:
    ```go
    type PDFConfig struct {
        FontSizes   map[string]float64  // "header": 16, "name": 20, "body": 10, etc.
        RowHeights  map[string]float64  // "header": 15, "summary": 15, "achievement": 6, etc.
        Margins     Margins             // Top, Bottom, Left, Right
        MinFontSize float64             // 8pt minimum for readability
    }
    ```
  - `DefaultConfig() PDFConfig` - returns default sizes from current main.go
  - `ScaleConfig(config PDFConfig, factor float64) PDFConfig` - scales all dimensions by factor

### Page Count Detection
- [ ] Research maroto page counting:
  - Check if maroto provides `GetPageCount()` after generation
  - Alternative: Save PDF to temp file, read page count with external library
  - Fallback: Use `github.com/pdfcpu/pdfcpu` to read page count from generated PDF

### Proportional Scaling Algorithm
- [ ] Implement scaling loop in `GeneratePDF`:
  ```go
  func GeneratePDF(resume Resume, outputPath string) error {
      config := DefaultConfig()

      for attempt := 1; attempt <= 3; attempt++ {
          // Generate PDF with current config
          pdf := buildPDFWithConfig(resume, config)
          pdf.SaveFile(outputPath)

          // Check page count
          pageCount := getPageCount(outputPath)
          if pageCount <= 1 {
              return nil  // Success!
          }

          // Scale down for next attempt
          scaleFactor := 1.0 - (float64(attempt) * 0.05)  // 0.95, 0.90, 0.85
          config = ScaleConfig(config, scaleFactor)

          // Ensure minimum font size
          if config.FontSizes["body"] < config.MinFontSize {
              return fmt.Errorf("cannot fit on 1 page while maintaining readability")
          }
      }

      return fmt.Errorf("exceeded max scaling attempts")
  }
  ```

### Overflow Handling
- [ ] Keep `overflow: true` flag in Achievement/Skill structs
- [ ] LLM marks longer text items with overflow flag
- [ ] Apply extra spacing for overflow items (scales proportionally too):
  - Normal achievement: 6 → scaled
  - Overflow achievement: 8 → scaled

---

## Phase 7: LLM Optimization

### Gemini Configuration
- [ ] Use Gemini 1.5 Flash (free tier):
  - 15 requests/min
  - 1M tokens/min
  - 1,500 requests/day
- [ ] Add retry logic with exponential backoff:
  - Handle rate limits (429)
  - Retry transient errors (500, 503)
  - Max 3 retries
- [ ] Log token usage:
  - Input tokens
  - Output tokens
  - Request count
  - Track against daily limit

### Prompt Templates
- [ ] Create `internal/llm/prompts.go`:
  - `ResumeGenerationPrompt(experience, jobInfo, attempt)`
  - `CoverLetterPrompt(resume, jobInfo)`
  - `UniversalInfoPrompt(resume, jobInfo)`
  - `FieldFillingPrompt(label, fieldType, options, context)`
- [ ] Use system/user message format:
  ```
  System: "You are a professional resume writer..."
  User: "{experience content}\n\nJob: {job_description}\n\nGenerate resume JSON."
  ```

### Context Window Management
- [ ] Limit experience.md size if too large (>10k chars):
  - Truncate oldest/least relevant entries
  - Keep most recent 2-3 experiences + most impressive projects
- [ ] For field filling, only send relevant context:
  - Resume summary (not full resume)
  - Relevant skills
  - Cover letter first paragraph
  - Universal info for that question type

---

## Implementation Order

1. **Phase 1**: Resume generation pipeline (LLM → JSON → PDF)
2. **Phase 6**: 1-page constraint enforcement
3. **Phase 2**: Form autofill (constants → LLM fallback)
4. **Phase 5**: API endpoints
5. **Phase 3**: Google Docs tracking
6. **Phase 4**: File management
7. **Phase 7**: LLM optimization

---

## Testing Plan

### Manual Testing
- [ ] Test resume generation with sample job descriptions
  - Verify 1-page output
  - Check PDF formatting (maroto rendering)
  - Validate overflow handling
- [ ] Test form filling with real job application sites:
  - Indeed
  - LinkedIn
  - Greenhouse
  - Workday
  - Lever
- [ ] Test edge cases:
  - Very long job descriptions
  - Multiple select fields
  - Radio button groups
  - Required vs optional fields

### Metrics to Track
- [ ] Resume generation:
  - Success rate (1 page on first try)
  - Retry count
  - Token usage per generation
- [ ] Form filling:
  - Constant match rate (target: 60%+)
  - LLM fallback rate (target: 40%-)
  - Field fill accuracy (manual verification)
- [ ] Performance:
  - Resume generation time (target: <10s)
  - Form fill time (target: <5s)
  - API response time

---

## Key Differences from Previous Design

❌ **Removed**: SQLite database, field mappings table, caching, applications table
❌ **Removed**: Admin bookmarklet, CRUD interface
❌ **Removed**: Multiple roles (fullstack/frontend/qa) - generate fresh each time
❌ **Removed**: Static/role-specific/dynamic field type classification
❌ **Removed**: Web scraping for job descriptions
❌ **Removed**: LLM regeneration loops for content trimming

✅ **Added**: experience.md as single source of truth
✅ **Added**: Fresh LLM generation per job application
✅ **Added**: Single LLM call with strict content constraints (1-2 positions, max 3 bullets each)
✅ **Added**: Proportional PDF scaling if content overflows (scales fonts/spacing, not content)
✅ **Added**: Manual job description input (copy/paste from LinkedIn/Indeed)
✅ **Added**: Google Docs tracking instead of database
✅ **Added**: Simple constants.json for basic fields
✅ **Added**: Cover letter + universal application info generation
✅ **Added**: Session-based caching (in memory, 1 hour TTL)

---

## File Structure

```
autofill-api/
├── experience.md                    # Source of truth (from resume-builder)
├── constants.json                   # Static autofill values
├── main.go                          # HTTP server + endpoints
├── .env                            # GEMINI_API_KEY, GOOGLE_DOC_ID
├── fonts/                          # DejaVu fonts for PDF
│   ├── DejaVuSans.ttf
│   ├── DejaVuSans-Bold.ttf
│   └── ...
├── icons-png/                      # Icons for PDF
│   ├── map-pin.png
│   ├── phone.png
│   └── ...
├── generated/                      # Output folder
│   └── {date}_{company}_{role}/
│       ├── resume.json
│       ├── resume.pdf
│       ├── cover_letter.txt
│       └── application_info.json
├── internal/
│   ├── input/
│   │   └── parser.go               # Parse experience.md + job description
│   ├── generator/
│   │   ├── resume_generator.go     # LLM → resume.json
│   │   ├── pdf_generator.go        # JSON → PDF (maroto)
│   │   ├── cover_letter_generator.go
│   │   ├── application_info.go
│   │   └── page_estimator.go       # 1-page constraint logic
│   ├── constants/
│   │   └── loader.go               # Load constants.json
│   ├── matcher/
│   │   ├── matcher.go              # Constant field matching
│   │   └── llm_matcher.go          # LLM fallback for unknown fields
│   ├── storage/
│   │   └── files.go                # File management
│   ├── tracking/
│   │   └── google_docs.go          # Append to Google Doc
│   └── llm/
│       ├── client.go               # Gemini API client
│       └── prompts.go              # Prompt templates
└── public/
    └── script.js                   # Bookmarklet (enhanced field collection)
```

---

## Next Steps

Start with Phase 1: Get the core resume generation working (experience.md → LLM → resume.json → PDF with 1-page constraint).
