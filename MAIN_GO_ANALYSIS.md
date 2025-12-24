# main.go Analysis & Proposed Changes

## What main.go Currently Does

### Server Setup
- Loads `.env` file for environment variables (GEMINI_API_KEY, PORT, IP)
- Loads job context from `job-description.txt` on startup
  - Line 1: Company name
  - Line 2: Job title
- Starts HTTPS server on 0.0.0.0:8000 (or custom PORT)
- Uses `cert.pem` and `key.pem` for SSL

### API Endpoints

#### 1. `/script.js` (GET)
- Serves the bookmarklet script from `public/script.js`
- Injects `API_URL` constant at the top of the script
- Sets CORS headers to allow cross-origin requests

#### 2. `/api/fill` (POST)
**Request format:**
```json
{
  "fields": [array of field objects],
  "job_context": { "title": "...", "company": "...", "url": "..." },
  "constants_only": bool
}
```

**Current Logic:**
1. Loads constants from `constants.json`
2. Loads company info from `company-info.txt` (if exists)
3. **Step 1:** Try to match ALL fields with constants using fuzzy matching
   - Matched fields → add to response
   - Unmatched fields → save for LLM
4. **Step 2:** If `constants_only` is false AND there are unmatched fields:
   - Send each unmatched field to LLM (Gemini API)
   - Get LLM response for each field
   - Add LLM results to response
5. Saves response to `responses/response_TIMESTAMP.json`
6. Returns combined results

**Response format:**
```json
{
  "fields": { "field_id": "value", ... },
  "metadata": {
    "constant_matches": 5,
    "llm_matches": 3,
    "total_fields": 8
  }
}
```

#### 3. `/api/context` (GET)
- Returns current job context (loaded from job-description.txt)

#### 4. `/api/constants` (GET/POST)
- **GET:** Returns all constants from `constants.json`
- **POST:** Updates `constants.json` with new values

#### 5. `/api/recent` (GET)
- Finds most recent file in `responses/` directory
- Returns that response JSON

---

## Problems with Current Implementation

### 1. No "LLM Only" Mode
- Current modes:
  - `constants_only: true` → only constants
  - `constants_only: false` → constants + LLM for unmatched
- **Missing:** Send ONLY specific fields to LLM (ignore constants entirely)

### 2. Always Processes All Fields
- Frontend sends ALL empty fields to `/api/fill`
- Backend tries to match all of them with constants first
- **New workflow needs:** Frontend sends only fields marked with "llm-fill"

### 3. No Way to Skip Constants Matching
- If user has already filled constants, LLM call still tries constants first
- Wastes processing time and makes logs messy

---

## Proposed Changes to main.go

### Separate Endpoints (CLEANEST APPROACH)

**Instead of boolean flags, create two separate endpoints:**

#### `/api/fill-constants` (POST)
**Purpose:** Only fuzzy match constants, no LLM

**Request:**
```json
{
  "fields": [array of field objects]
}
```

**Logic:**
1. Load constants from `constants.json`
2. Fuzzy match fields to constants
3. Return matched results

**Response:**
```json
{
  "fields": { "field_id": "value", ... },
  "metadata": {
    "matched": 5,
    "total_fields": 8
  }
}
```

---

#### `/api/fill-llm` (POST)
**Purpose:** Only send to LLM, skip constants entirely

**Request:**
```json
{
  "fields": [array of field objects],
  "job_context": { "title": "...", "company": "...", "url": "..." }
}
```

**Logic:**
1. Load job context from `job-description.txt`
2. Load company info from `company-info.txt`
3. Load experience from `experience.txt`
4. Send each field to LLM (Gemini API)
5. Return LLM results
6. Save response to `responses/response_TIMESTAMP.json`

**Response:**
```json
{
  "fields": { "field_id": "value", ... },
  "metadata": {
    "llm_matches": 3,
    "total_fields": 3
  }
}
```

---

### Why This is Better

**Pros:**
- Single responsibility per endpoint
- No confusing boolean logic
- Clear naming
- Easier to debug
- Easier to test
- Can keep old `/api/fill` endpoint for backward compatibility

**Cons:**
- None (this is objectively better)

---

## Frontend Changes

**Fill Constants Button:**
```js
POST /api/fill-constants
{
  fields: [all empty fields]
}
```

**Fill LLM Button:**
```js
POST /api/fill-llm
{
  fields: [only fields where value === "llm-fill"],
  job_context: { ... }
}
```

---

## Other Cleanup Opportunities

### 1. Remove `/api/recent` endpoint?
- Script.js has a "Fill from Recent" button but user wants to remove it
- If removing button, can remove endpoint too

### 2. Better Error Messages
- Currently just logs to console
- iOS has no console access
- Could return errors in response for UI to display

### 3. Company Info Loading
- Currently loads `company-info.txt` on every `/api/fill` request
- Could load once on startup like job-description.txt
- Or combine both files into one

### 4. Response Saving
- Saves every response to disk in `responses/`
- Could add option to disable this
- Could add cleanup for old files

---

## What NOT to Change

- Don't touch the fuzzy matching logic in `internal/matcher`
- Don't touch LLM integration in `internal/llm`
- Don't change constants loading in `internal/constants`
- Keep all existing endpoints (backward compatibility)
- Keep CORS headers as-is
- Keep HTTPS setup as-is

---

## Summary

**Minimal change needed:**
1. Add `LLMOnly bool` field to FillRequest
2. Add conditional in handleFill to skip constants when `LLMOnly == true`
3. Update logs to show mode being used

**Frontend will handle:**
- Scanning page for "llm-fill" values
- Sending only those fields with `llm_only: true` flag
