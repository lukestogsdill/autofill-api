# Autofill API - Refactor TODO

## Overview
iOS bookmarklet tool for autofilling job application forms using constants + LLM for specific fields.

## Current Architecture

### Frontend (Bookmarklet)
- `public/script.js` - Main bookmarklet script
- Opens modal menu with 3 functions:
  1. Fill Constants (fuzzy match saved data to form fields)
  2. Fill LLM (scan for fields marked "llm-fill", send to API)
  3. Settings (manage constants)

### Backend (API)
- Constants storage (GET/POST `/api/constants`)
- LLM fill endpoint (POST `/api/fill`)
- Job context endpoint (GET `/api/context`)
- Recent response cache (GET `/api/recent`)

## Workflow

1. User opens job application form
2. Clicks bookmarklet → modal menu appears
3. User clicks "Fill Constants"
   - Script collects all form fields
   - Fuzzy matches field labels/names/placeholders to saved constants
   - Auto-fills matching fields
4. User manually types "llm-fill" in any remaining empty fields they want AI to fill
5. User clicks "Fill LLM"
   - Script scans page for fields with value === "llm-fill"
   - Collects only those fields
   - Sends to LLM API with field metadata (label, name, type, etc.)
   - API returns generated content
   - Script fills those fields with LLM responses

## Refactor Goals

### Structure
- [x] Separate concerns into modules
- [x] Extract all styles to STYLES object
- [x] Create HTML templates as functions
- [ ] Clean up and simplify logic flow
- [ ] Better error handling
- [ ] Better debug logging

### LLM Trigger Change
- [ ] Change from constants-based trigger to form-value-based trigger
- [ ] Scan form fields for value === "llm-fill"
- [ ] Only send marked fields to LLM API (not all empty fields)
- [ ] Clear "llm-fill" text before filling with LLM response

### Menu Updates
- [ ] Update button labels
  - "Fill Constants" (step 1)
  - "Fill LLM Fields" (step 2)
  - "Settings" (manage constants)
- [ ] Add workflow instructions in menu
- [ ] Remove "Fill All" and "Fill Recent" buttons (keep it simple)

### Settings
- [ ] Keep current constants management
- [ ] Remove "llm-fill" placeholder from constants settings
- [ ] Just allow adding/editing/removing constant key-value pairs

## Open Questions

- [ ] Google Doc integration - needed or scrap?
//need
  - What data to save?
  - When to save?
  - What API endpoint?
  //dont worry for now

- [ ] Job context for LLM
  - Does LLM need job title/company/description?
  //company name on line 1 and title on line 2 on job-description.txt
  - Where does this context come from?
  job-description, company-info, and experience
  - Manual entry? Scrape from page?
ill manually enter
- [ ] Fuzzy matching threshold
  - Current threshold: 0.5
  - Need to test and adjust?

- [ ] Error handling
  - What if no forms found?
  - What if API fails?
  - What if no fields marked for LLM?
  //this one is tricky bc i dont have access to dev tools. maybe some way to view logs on menu

## Implementation Order

1. **Phase 1: Core Refactor**
   - Keep existing functionality
   - Clean up structure (already mostly done)
   - Test that it still works

2. **Phase 2: LLM Trigger Change**
   - Implement form-value scanning for "llm-fill"
   - Update Fill LLM button to use new logic
   - Test LLM fill workflow

3. **Phase 3: Menu/UI Updates**
   - Update button labels and workflow
   - Remove unused buttons
   - Add instructions

4. **Phase 4: Polish**
   - Error handling
   - Debug logging
   - Edge cases
   - Testing

5. **Phase 5: Additional Features (TBD)**
   - Google Doc integration?
   - Job context improvements?
   - Other enhancements?

## Notes

- iOS bookmarklet restrictions: must work with single file injection
- Fuzzy matching is key - field names vary wildly across sites
- LLM responses need job context for quality (title, company, description)
- User manually marks fields for LLM = precise control over API costs
