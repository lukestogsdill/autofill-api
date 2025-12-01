# Autofill API - Simplified Implementation

## Architecture Overview

**Goal:** Match form field labels to your answer keys, using field-level caching + LLM fallback.

**Data Structure:**
```
data/
  ├── field_mappings.json    # label→key cache (learned over time)
  ├── user_data.json          # your actual answers (short ones)
  └── content/                # long answers in separate files
      ├── cover_letter.txt
      ├── why_this_company.txt
      └── project_description.txt
```

**Flow:**
1. Form fields come in → normalize labels
2. Check `field_mappings.json` for cached matches
3. For uncached fields → send to Gemini with context
4. Cache new mappings → return values from `user_data.json`

---

## Phase 1: Data Storage Setup (SQLite)

### Database Schema
- [ ] Create SQLite database `data.db`
- [ ] Create tables:
  ```sql
  -- Field mappings: label variants → user_data keys (shared across roles)
  CREATE TABLE field_mappings (
    id INTEGER PRIMARY KEY,
    user_data_key TEXT NOT NULL,
    label_variant TEXT NOT NULL,
    confidence REAL DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_label ON field_mappings(label_variant);

  -- User data: shared + role-specific values
  CREATE TABLE user_data (
    id INTEGER PRIMARY KEY,
    key TEXT NOT NULL,
    role TEXT,  -- NULL = shared, 'fullstack'/'frontend'/'qa' = role-specific
    value TEXT,
    is_file_reference BOOLEAN DEFAULT 0,
    UNIQUE(key, role)
  );

  -- Settings: active role selection
  CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT
  );
  ```

- [ ] Seed initial data:
  ```sql
  -- Shared fields
  INSERT INTO user_data (key, role, value) VALUES
    ('first_name', NULL, 'Luke'),
    ('last_name', NULL, 'Henderson'),
    ('email', NULL, 'luke@example.com'),
    ('phone', NULL, '555-123-4567');

  -- Role-specific fields
  INSERT INTO user_data (key, role, value, is_file_reference) VALUES
    ('cover_letter', 'fullstack', '@file:content/cover_fs.txt', 1),
    ('cover_letter', 'frontend', '@file:content/cover_fe.txt', 1),
    ('cover_letter', 'qa', '@file:content/cover_qa.txt', 1);

  -- Initial mappings
  INSERT INTO field_mappings (user_data_key, label_variant) VALUES
    ('first_name', 'first name'),
    ('first_name', 'given name'),
    ('first_name', 'fname'),
    ('email', 'email'),
    ('email', 'email address');

  -- Default active role
  INSERT INTO settings (key, value) VALUES ('active_role', 'fullstack');
  ```

- [ ] Create `data/content/` directory for long answers (cover letters, etc.)

### Go Data Layer
- [ ] Create `internal/storage/db.go`:
  - `InitDB()` - initialize SQLite connection, create tables if needed
  - `GetFieldMapping(label) (key, found)` - lookup cached mapping
  - `SaveFieldMapping(key, label, confidence)` - learn new mapping
  - `GetUserData(key, role) (value)` - get value with role fallback (role-specific → shared)
  - `SetUserData(key, role, value)` - update user data
  - `GetActiveRole()` - get current role from settings
  - `SetActiveRole(role)` - change active role
  - `LoadFileContent(path)` - handle @file: references
  - `ListAllMappings()` - for admin UI
  - `ListUserData(role)` - for admin UI

---

## Phase 2: Field Matching Logic

### Core Matcher
- [ ] Create `internal/matcher/matcher.go`:
  - `MatchField(label, fieldType, context) (key, found)`:
    1. Normalize label (lowercase, trim, remove special chars)
    2. Check cache (labelToKey map)
    3. Return key if found
  - Handle multi-value fields:
    - "firstname, lastname, email of referral one" → could map to multiple keys
    - Comma-separated = multiple keys

### LLM Integration (Gemini)
- [ ] Create `internal/llm/gemini.go`:
  - `ClassifyField(label, fieldType, context, availableKeys) (key string, confidence float64)`
  - Prompt design:
    ```
    You are mapping form field labels to user data keys.

    Available keys: first_name, last_name, email, phone, reference1_name, reference1_email, ...

    Field label: "Please provide reference's first name"
    Field type: text
    Context: Previous field was "Can we contact references?" (checkbox)

    Return ONLY the matching key name. If the field asks for multiple pieces of info (comma-separated), return comma-separated keys.
    ```
  - Use Gemini Flash (cheapest model)
  - Add retry logic with exponential backoff
  - Log token usage for monitoring

### Smart Filling
- [ ] Create `internal/filler/filler.go`:
  - `FillField(field Field, userDataKey string) (value string, err error)`
  - Handle field types:
    - Text/textarea: return value directly
    - Select dropdown: LLM picks best option from available options
    - Radio: same as select
    - Checkbox: convert boolean/yes/no to checked state
    - Date: format conversion if needed

---

## Phase 3: API Implementation

### Update `/api/fill` Handler
- [ ] Get active role from DB
- [ ] Parse incoming fields from bookmarklet
- [ ] For each field:
  1. Try cache lookup (field_mappings table)
  2. If not cached → call LLM
  3. Save new mapping to DB
  4. Get value from user_data (with role fallback)
  5. Fill appropriately based on field type
- [ ] Return filled values:
  ```json
  {
    "field1_name": "Luke",
    "field2_name": "Henderson",
    "field3_name": "luke@example.com",
    "metadata": {
      "role": "fullstack",
      "cache_hits": 8,
      "llm_calls": 2
    }
  }
  ```
- [ ] Handle errors gracefully (partial fills OK)

### Admin API Endpoints
- [ ] `GET /api/settings` - get active role & other settings
- [ ] `POST /api/settings` - update active role
- [ ] `GET /api/user-data?role=fullstack` - list all user data for a role
- [ ] `POST /api/user-data` - create/update user data entry
  ```json
  { "key": "cover_letter", "role": "frontend", "value": "...", "is_file_reference": false }
  ```
- [ ] `DELETE /api/user-data` - delete user data entry
- [ ] `GET /api/mappings` - list all field mappings
- [ ] `POST /api/mappings` - add new mapping
  ```json
  { "user_data_key": "first_name", "label_variant": "given name" }
  ```
- [ ] `DELETE /api/mappings/:id` - remove incorrect mapping

### Context Awareness
- [ ] Send surrounding fields to LLM for context:
  - Include previous 2 fields + next 2 fields in prompt
  - Include section headings if available
  - Include field grouping info (fieldset/legend)
- [ ] Detect multi-part fields:
  - "firstname, lastname, email of referral one"
  - Return multiple keys: `["reference1_name", "reference1_last_name", "reference1_email"]`

---

## Phase 4: Enhanced Field Detection (script.js)

### Expand Field Collection
- [ ] Add support for:
  - Select dropdowns (include available options)
  - Radio buttons (group by name, include options)
  - Checkboxes (distinguish single vs multi-select)
  - Textareas
  - Date inputs
- [ ] Extract better context:
  - Fieldset/legend text
  - Section headings (h2, h3 above field)
  - Placeholder text
  - Help text / aria-labels
- [ ] Send field structure:
  ```javascript
  {
    label: "First Name",
    name: "fname",
    type: "text",
    required: true,
    context: "Personal Information section",
    placeholder: "Enter your first name"
  }
  ```

---

## Phase 5: Admin Bookmarklet (admin.js)

### Bookmarklet Creation
- [ ] Create `admin.js` - CRUD interface injected as modal overlay
- [ ] Serve at `/admin.js` endpoint (like script.js)
- [ ] Create bookmarklet code:
  ```javascript
  javascript:(function(){var s=document.createElement('script');s.src='https://YOUR_IP:8000/admin.js';document.body.appendChild(s);})();
  ```

### Admin UI Features
- [ ] Inject modal overlay on current page (fixed position, high z-index)
- [ ] Role selector dropdown:
  - Fetch current role from `/api/settings`
  - Switch between fullstack/frontend/qa
  - POST to `/api/settings` on change
- [ ] User Data Section:
  - Fetch from `/api/user-data?role={activeRole}`
  - Show shared fields (role=NULL) in one section
  - Show role-specific fields in another section
  - Inline edit functionality
  - Add new key/value pairs
  - Delete entries
  - Handle @file: references (show indicator, allow editing)
- [ ] Field Mappings Section:
  - Fetch from `/api/mappings`
  - Group by user_data_key
  - Show all label variants for each key
  - Add new label variants
  - Delete incorrect mappings
- [ ] Stats display:
  - Total mappings cached
  - Fields per role
  - Recent LLM calls (if tracked)

### UI Design
- [ ] Clean, mobile-friendly modal
- [ ] Close button (X in corner)
- [ ] Tabbed interface: Settings | User Data | Mappings
- [ ] Save buttons with loading states
- [ ] Success/error notifications
- [ ] Dark overlay backdrop

### Token Optimization
- [ ] Log metrics:
  - Cache hit rate per form
  - Number of LLM calls per form
  - Token usage per request
- [ ] Target metrics:
  - Form 1: ~10-20 LLM calls
  - Form 5: ~2-3 LLM calls
  - Form 10: ~0-1 LLM calls
- [ ] Batch uncached fields in single LLM request:
  - Send all uncached fields together for context
  - Get all mappings back at once

---

## Phase 6: Edge Cases & Polish

### Handle Tricky Scenarios
- [ ] Multi-value fields (comma-separated)
- [ ] Conditional fields (only shown if X is selected)
- [ ] Date format conversions
- [ ] Phone number formatting
- [ ] Address fields (street, city, state, zip as separate or combined)
- [ ] Arrays of references/jobs (reference1, reference2, reference3)

### Response Format
- [ ] Include confidence scores:
  ```json
  {
    "fields": {
      "fname": "Luke",
      "email": "luke@example.com"
    },
    "metadata": {
      "cache_hits": 8,
      "llm_calls": 2,
      "confidence": {
        "fname": 1.0,      // cached
        "email": 1.0,      // cached
        "fallback_field": 0.7 // LLM guessed
      }
    }
  }
  ```

---

## Current Status

- ✅ Basic HTTPS API server
- ✅ Bookmarklet injection
- ✅ Form field collection (text inputs only)
- ⏸️ Returns dummy data (hardcoded)
- ❌ No data storage
- ❌ No field mappings
- ❌ No LLM integration
- ❌ Limited field type support

---

## Implementation Order

1. **Phase 1**: Set up SQLite database + storage layer
2. **Phase 2**: Build matcher with cache lookup + Gemini integration
3. **Phase 3**: Update `/api/fill` handler + add admin API endpoints
4. **Phase 4**: Enhance script.js field detection
5. **Phase 5**: Build admin.js bookmarklet for CRUD operations
6. **Phase 6**: Handle edge cases as you encounter them

---

## Key Benefits of This Architecture

✅ **SQLite over JSON**: Scales better, handles concurrent writes, enables fuzzy search
✅ **Role-based data**: One DB, multiple job types, shared fields reduce duplication
✅ **Bookmarklet-based admin**: Perfect for iOS, no need for separate app
✅ **LLM + caching**: Expensive on first form, nearly free after learning
✅ **Progressive enhancement**: Each phase builds on the previous, all independently useful


