# Autofill API

A smart autofill system for job applications using semantic matching and AI. Works automatically on any website with forms via iOS Safari userscript.

## Features

- 🚀 **Auto-loads on form pages** - No bookmarks to click
- 🧠 **Semantic matching** - Understands field variations using embeddings
- 🤖 **AI fallback** - LLM generates custom answers for complex questions
- 📱 **iOS optimized** - Works seamlessly in Safari with Userscripts extension
- ⚡ **Fast & cheap** - Caches embeddings, only uses LLM when needed

## Quick Start

### 1. Server Setup

Generate SSL certificate:
```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
  -subj "/CN=YOUR_SERVER_IP" -addext "subjectAltName=IP:YOUR_SERVER_IP"
```

Create `.env` file:
```bash
IP=YOUR_SERVER_IP
PORT=4444
GEMINI_API_KEY=your_api_key_here
```

Install dependencies and run:
```bash
go mod download
go run .
```

Server starts on `https://0.0.0.0:4444`

### 2. iOS Safari Setup

**Install Userscripts Extension:**
1. Download **Userscripts** from App Store (free)
2. Settings → Safari → Extensions → Userscripts → **ON**

**Install the Script:**
1. Visit `https://YOUR_SERVER_IP:4444/autofill.user.js` in Safari
2. Accept certificate warning (one-time)
3. Tap **Share** → **Userscripts** → **Install Script**

**Done!** The ✨ button now appears automatically on any page with forms.

## Usage

### Workflow:

1. Visit any job application page
2. ✨ floating button appears automatically
3. Tap it to open the menu
4. **Fill Constants** - Fills name, email, phone, etc. instantly
5. **Type "##"** in empty fields you want AI to complete
6. **Fill LLM Fields** - AI writes custom responses for marked fields
7. Review and submit!

### How Matching Works:

```
Field: "Are you authorized to work in the US?"
  ↓
1. Exact match: ❌ (not literally "authorized_to_work")
2. Semantic match: ✅ (87% similarity to "authorized to work")
   → Returns: constants["authorized_to_work"] = "yes"
3. Negation check: ✅ (no negation words)
4. ✓ Field filled!
```

**Matching Strategy:**
- **Exact** → Fast lookup for known field names
- **Semantic** → Embeddings compare meaning (threshold: 0.7)
- **Fuzzy** → Pattern matching fallback
- **LLM** → Only for truly complex/unique fields

## Configuration

Edit `constants.json` to customize your data:
```json
{
  "first_name": "Luke",
  "email": "you@example.com",
  "authorized_to_work": "yes",
  "years_experience": "5"
}
```

Or edit via the ⚙️ settings menu in the app.

## Supported Sites

**All of them!** The userscript runs on every website (`@match *://*/*`)

Tested on:
- Greenhouse, Lever, Workday, Ashby, BambooHR
- Google Forms
- Custom application forms
- Any website with `<input>` fields

## API Endpoints

- `GET /autofill.user.js` - Userscript installer
- `GET /script.js` - Main autofill script
- `POST /api/fill-constants` - Fill with constants only
- `POST /api/fill-llm` - Fill marked fields with AI
- `GET /api/constants` - Get constants
- `POST /api/constants` - Update constants

## Architecture

```
User visits form page
  ↓
Userscript detects forms → Injects script.js
  ↓
1. Fill Constants (semantic matching)
   - Batch embed constant keys at startup
   - Cosine similarity: field vs cached vectors
   - Threshold 0.7 → instant match
  ↓
2. User marks fields with "##"
  ↓
3. Fill LLM (only for marked fields)
   - Send to Gemini Flash Lite
   - Context: job description, resume, constants
   - Generate personalized answer
  ↓
4. Done!
```

## Security

- **Self-signed SSL** - Accept certificate warning on first visit
- **HTTPS required** - Modern browsers block mixed content
- **Local server** - Data never leaves your device/network
- **API key** - Store Gemini API key in `.env` file

## Troubleshooting

**Button doesn't appear:**
- Check Userscripts extension is enabled in Safari settings
- Look for blue notification in top-right when page loads
- Open Safari console (aA → Web Inspector) and check for `[Autofill]` logs

**Certificate errors:**
- Visit the userscript URL directly once and accept the warning
- Certificate is valid for 365 days from creation

**Semantic matcher failed:**
- Check `GEMINI_API_KEY` is set in `.env`
- Verify API key has embedding access
- Falls back to fuzzy matching if embeddings fail

**Forms not filling:**
- Check browser console for errors
- Verify server is running and accessible
- Try manual fill using the menu buttons

## Files

- `public/autofill.user.js` - iOS Safari userscript
- `public/script.js` - Main autofill logic
- `internal/matcher/semantic_matcher.go` - Embedding-based matching
- `internal/matcher/matcher.go` - Field matching logic
- `internal/matcher/llm_matcher.go` - AI field filling
- `constants.json` - Your personal data
- `IOS_SETUP.md` - Detailed iOS setup guide

## License

MIT
