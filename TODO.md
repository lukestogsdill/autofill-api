# Autofill API - TODO

## Workflow Overview
```
1. Input: experience.md + job description + constants.json
2. Generate: Resume PDF + Cover Letter
3. Bookmarklet: Form replica modal → Selective field filling
4. Track: Update Google Doc with application record
```

**Constraints**: Must work on iPhone browser (Safari) using bookmarklets only.

---

## NEW MODAL ARCHITECTURE & IMPLEMENTATION PLAN

### Overview
Complete redesign of the bookmarklet interface to display a form replica inside the modal, with selective field filling and smart constant management.

---

### Part 1: State Management & Caching

**LocalStorage Schema:**
```javascript
{
  "autofill_constants": {
    "first_name": "Luke",
    "disability_status": "no",
    "veteran_status": "no",
    // ... all constants
  },
  "autofill_constants_timestamp": 1702847123456,
  "autofill_script_cache": "...(full script.js code)...",
  "autofill_script_timestamp": 1702847123456,
  "autofill_form_state": {
    "linkedin.com/jobs/123": {
      "selections": ["email", "phone", "why_work_here"],
      "values": {
        "email": "luke@...",
        "phone": "(832)...",
        "why_work_here": "I'm excited..."
      }
    }
  }
}
```

**Caching Strategy:**
- Constants: 1 hour TTL
- Script: 5 minute TTL
- Form state: Per-URL (persist selections between reopens)

**Implementation:**
```javascript
class CacheManager {
  static TTL_CONSTANTS = 3600000; // 1 hour
  static TTL_SCRIPT = 300000;     // 5 minutes

  getConstants() {
    // Check timestamp, fetch if expired
    // Return from localStorage if valid
  }

  saveConstants(data) {
    // Save with current timestamp
  }

  getFormState(url) {
    // Load saved selections for this URL
  }
}
```

---

### Part 2: New Modal UI Structure

**Visual Layout:**
```
┌─────────────────────────────────────────────┐
│  ✕                                   ⚙️      │  <- Fixed header
│  [Config Fill] [LLM Fill]                   │  <- Action buttons
│  [Select All] [Clear]                       │  <- Selection helpers
├─────────────────────────────────────────────┤
│                                              │
│  ☐ First Name                               │  <- Scrollable form
│    [Luke                          ]         │     replica area
│                                              │
│  ☑ Email                                    │  <- Checked field
│    [lukestogsdill@gmail.com       ]         │
│                                              │
│  ☐ Why do you want to work here?           │
│    [________________________________]       │
│    [________________________________]       │
│                                              │
│  ☐ Years of Experience                      │
│    [1-3                        ▼]           │  <- Dropdown
│                                              │
│  ☑ Are you a veteran?                       │  <- Checked field
│    ○ Yes  ● No                              │  <- Radio buttons
│                                              │
├─────────────────────────────────────────────┤
│  [Save Selected as Constants]               │  <- Bottom action
└─────────────────────────────────────────────┘
```

**Components:**
1. **Header Bar** (fixed, 60px)
   - Left: X button to close
   - Right: ⚙️ settings cog (opens constants manager)

2. **Action Bar** (fixed, 50px)
   - Config Fill button (blue gradient)
   - LLM Fill button (green gradient)

3. **Selection Bar** (fixed, 40px)
   - Select All (small button)
   - Clear (small button)
   - Counter: "3 of 12 selected"

4. **Form Area** (scrollable)
   - Each field rendered as:
     ```html
     <div class="field-row">
       <input type="checkbox" class="field-select" />
       <div class="field-content">
         <label>Field Label</label>
         <input/select/textarea> <!-- Actual input -->
       </div>
     </div>
     ```

5. **Bottom Actions** (fixed, 60px)
   - "Save Selected as Constants" button

---

### Part 3: Semantic Constant Matching System

**The Problem:**
Questions are worded differently but mean the same thing:
- "Are you a veteran?" vs "Do you have military service?"
- "Are you disabled?" vs "Do you NOT have a disability?" (negation!)

**The Solution:**
Store constants by **semantic meaning**, not literal wording.

**Semantic Constant Categories:**
```javascript
const SEMANTIC_CATEGORIES = {
  // Personal Info
  identity: {
    patterns: ['first name', 'fname', 'given name'],
    constant_key: 'first_name'
  },

  // Yes/No Status Questions
  disability: {
    constant_key: 'disability_status',
    type: 'boolean'
  },

  veteran: {
    patterns: ['veteran', 'military', 'armed forces', 'service member'],
    constant_key: 'veteran_status',
    type: 'boolean'
  },

  sponsorship: {
    patterns: ['sponsorship', 'visa', 'work authorization', 'h1b'],
    constant_key: 'require_sponsorship',
    type: 'boolean'
  },

  felony: {
    patterns: ['convicted', 'felony', 'criminal', 'crime'],
    constant_key: 'convicted_felony',
    type: 'boolean'
  },

  clearance: {
    patterns: ['security clearance', 'clearance', 'classified'],
    constant_key: 'security_clearance',
    type: 'boolean'
  },

  relocation: {
    patterns: ['relocate', 'move', 'willing to move', 'relocation'],
    constant_key: 'willing_to_relocate',
    type: 'boolean'
  }
}
```

**Matching Algorithm:**
```javascript
function matchSemanticConstant(fieldLabel, constants) {
  // 1. Normalize label
  const normalized = normalizeLabel(fieldLabel); // lowercase, no punctuation

  // 2. Detect negation
  const hasNegation = detectNegation(normalized);
  // Looks for: "not", "don't", "do not", "are you NOT"

  // 3. Extract topic
  for (const [category, config] of Object.entries(SEMANTIC_CATEGORIES)) {
    for (const pattern of config.patterns) {
      if (normalized.includes(pattern)) {
        // Found topic!
        const constantKey = config.constant_key;
        let value = constants[constantKey];

        // 4. Handle negation
        if (hasNegation && config.type === 'boolean') {
          value = invertBoolean(value); // "yes" -> "no", "no" -> "yes"
        }

        // 5. Map to field options
        if (field.options) {
          return matchValueToOptions(value, field.options);
        }

        return { value, matched: true, source: 'semantic' };
      }
    }
  }

  // 6. Fallback to direct matching
  return directMatch(fieldLabel, constants);
}
```

**Example Flows:**

**Example 1: Standard Question**
```
Field: "Are you a veteran?"
Options: ["Yes", "No"]

1. Normalize: "are you a veteran"
2. Detect negation: NO
3. Extract topic: "veteran" → constant_key: "veteran_status"
4. Get value: constants["veteran_status"] = "no"
5. Match to options: "no" → "No"
6. Return: "No"
```

**Example 2: Negated Question**
```
Field: "Are you NOT a person with a disability?"
Options: ["Yes", "No"]

1. Normalize: "are you not a person with a disability"
2. Detect negation: YES (contains "not")
3. Extract topic: "disability" → constant_key: "disability_status"
4. Get value: constants["disability_status"] = "no"
5. Invert due to negation: "no" → "yes"
6. Match to options: "yes" → "Yes"
7. Return: "Yes" (correct! You are NOT disabled)
```

**Example 3: Different Wording**
```
Field: "Do you require visa sponsorship?"
Options: ["I do", "I do not"]

1. Normalize: "do you require visa sponsorship"
2. Detect negation: NO
3. Extract topic: "sponsorship" → constant_key: "require_sponsorship"
4. Get value: constants["require_sponsorship"] = "no"
5. Match to options: "no" → "I do not"
6. Return: "I do not"
```

**Negation Detection:**
```javascript
function detectNegation(text) {
  const negationPatterns = [
    /\bnot\b/,
    /\bdon't\b/,
    /\bdo not\b/,
    /\bdoes not\b/,
    /\baren't\b/,
    /\bare not\b/,
    /\bisn't\b/,
    /\bis not\b/
  ];
  return negationPatterns.some(pattern => pattern.test(text));
}
```

**Value Mapping:**
```javascript
function matchValueToOptions(value, options) {
  const yesVariants = ['yes', 'y', 'true', '1', 'i do', 'i am', 'agree'];
  const noVariants = ['no', 'n', 'false', '0', 'i do not', 'i am not', 'disagree'];

  const normalizedValue = value.toLowerCase();

  for (const option of options) {
    const optionText = option.text.toLowerCase();

    if (yesVariants.includes(normalizedValue)) {
      if (yesVariants.some(v => optionText.includes(v))) {
        return option.value;
      }
    }

    if (noVariants.includes(normalizedValue)) {
      if (noVariants.some(v => optionText.includes(v))) {
        return option.value;
      }
    }
  }

  return value; // Fallback
}
```

---

### Part 4: Save as Constant Feature

**Workflow:**
1. User fills some fields (manually or via LLM)
2. Selects fields they want to save as constants
3. Clicks "Save Selected as Constants"
4. Modal shows each selected field with:
   - Current answer
   - Auto-suggested semantic name
   - Editable name input

**Example Modal:**
```
┌─────────────────────────────────────────┐
│  Save as Constants                      │
├─────────────────────────────────────────┤
│  Field: "Are you a veteran?"            │
│  Answer: "No"                           │
│  Save as: [veteran_status   ]          │  <- Editable
│           └─ suggested name             │
│                                         │
│  Field: "LinkedIn Profile URL"          │
│  Answer: "linkedin.com/in/luke..."      │
│  Save as: [linkedin_url      ]          │
│                                         │
│  [Cancel] [Save All]                    │
└─────────────────────────────────────────┘
```

**Auto-suggestion Logic:**
```javascript
function suggestConstantName(fieldLabel, answer) {
  // 1. Check if it matches a semantic category
  for (const [category, config] of SEMANTIC_CATEGORIES) {
    for (const pattern of config.patterns) {
      if (fieldLabel.toLowerCase().includes(pattern)) {
        return config.constant_key; // e.g., "veteran_status"
      }
    }
  }

  // 2. Generate from field label
  return fieldLabel
    .toLowerCase()
    .replace(/[^a-z0-9\s]/g, '')
    .replace(/\s+/g, '_')
    .slice(0, 30);
}
```

---

### Part 5: Selective Fill Workflow

**Config Fill (Constants Only):**
1. User checks fields they want to fill
2. Clicks "Config Fill"
3. System:
   - Loops through checked fields only
   - Tries semantic matching first
   - Falls back to direct matching
   - Fills matched fields
   - Shows: "Filled 5 of 7 selected fields"

**LLM Fill (Unmatched Fields):**
1. User checks fields they want to fill
2. Clicks "LLM Fill"
3. System:
   - Sends only checked fields to API
   - API tries constants first
   - Uses LLM for unmatched
   - Returns values for checked fields
   - Fills them in
   - Shows: "Filled 7 of 7 selected fields (2 constants, 5 LLM)"

---

### Part 6: Implementation Checklist

**Phase 1: Foundation**
- [ ] Create CacheManager class (localStorage wrapper)
- [ ] Implement constants caching with TTL
- [ ] Implement script caching with TTL
- [ ] Create state management for selections

**Phase 2: New Modal UI**
- [ ] Build new modal structure (header, action bar, form area, footer)
- [ ] Render form fields as replica with checkboxes
- [ ] Implement field selection (checkboxes)
- [ ] Add Select All / Clear buttons with counter
- [ ] Style with slate gradient theme

**Phase 3: Semantic Matching**
- [ ] Create SEMANTIC_CATEGORIES config
- [ ] Implement matchSemanticConstant()
- [ ] Implement detectNegation()
- [ ] Implement matchValueToOptions()
- [ ] Test with various question phrasings

**Phase 4: Selective Fill**
- [ ] Implement Config Fill (selected fields only)
- [ ] Implement LLM Fill (selected fields only)
- [ ] Update API to accept field selection
- [ ] Show fill results with counts

**Phase 5: Save as Constant**
- [ ] Build save-as-constant modal
- [ ] Implement suggestConstantName()
- [ ] Allow user to edit suggested names
- [ ] Save to constants.json via API
- [ ] Update cache

**Phase 6: Settings**
- [ ] Move constants manager to settings modal
- [ ] Add/remove constants functionality
- [ ] Cache invalidation on save

---

### Expected User Flow

1. **Open bookmarklet** → Modal shows form replica
2. **Auto-select common fields** → email, phone, name (pre-checked)
3. **Click "Config Fill"** → Instantly fills 8 fields from constants
4. **Review remaining fields** → Check ones they want LLM to fill
5. **Click "LLM Fill"** → AI fills 4 complex questions
6. **Manually edit one answer** → Adjust wording
7. **Select that field** → Click "Save as Constant"
8. **Save as "preferred_work_style"** → Now reusable
9. **Close modal** → All fields filled, ready to submit!

---

This architecture provides:
✅ Smart semantic matching (handles negation)
✅ Selective filling (only what you want)
✅ Persistent caching (fast loads)
✅ Constant building (learn over time)
✅ Full control (review before fill)

---

## Old Checklist (Reference - Will be replaced by above)

### 1. Resume Generation
- [x] Load experience.md
- [x] Parse job description from file
- [x] Generate resume JSON with LLM (Gemini)
- [x] Generate PDF from JSON (maroto)
- [ ] Refine LLM prompt for better resume output

### 2. Cover Letter Generation
- [ ] Create cover letter generator

### 3. Google Doc Tracking
- [ ] Set up Google Docs API integration
- [ ] Create tracking function
- [ ] Trigger after successful form fill
