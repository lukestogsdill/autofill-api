# Autofill API

A simple HTTPS API server that enables remote form autofilling via userscript (recommended for iOS) or bookmarklet.

## Setup

### 1. Generate SSL Certificates

On your server, run:

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=YOUR_SERVER_IP" -addext "subjectAltName=IP:YOUR_SERVER_IP"
```

Replace `YOUR_SERVER_IP` with your server's public IP address.

### 2. Configure Environment

Create a `.env` file:

```bash
IP=YOUR_SERVER_IP
PORT=8000
```

Replace `YOUR_SERVER_IP` with your server's public IP address.

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Server

```bash
go run main.go
```

The server will start on `https://0.0.0.0:8000`

## Usage

### Option 1: Userscript (RECOMMENDED FOR iOS)

**Why Userscript?** Auto-loads on job pages without clicking anything. Just visit the page and the autofill button appears!

#### Setup on iOS Safari:

1. **Install Userscripts Extension**
   - Download "Userscripts" from the iOS App Store (free)
   - Enable it in Safari: Settings → Safari → Extensions → Userscripts → On

2. **Trust the Certificate** (One-time)
   - Open Safari and visit `https://YOUR_SERVER_IP:PORT/autofill.user.js`
   - Accept the security warning (self-signed cert)
   - You should see the userscript code

3. **Install the Userscript**
   - In Safari, visit `https://YOUR_SERVER_IP:PORT/autofill.user.js`
   - Tap the "Share" button → "Userscripts" → "Install Script"
   - The script is now active!

4. **Use It**
   - Visit any job application page (Lever, Greenhouse, Workday, etc.)
   - The ✨ floating button automatically appears
   - Tap to open the autofill menu

**Supported Sites:** Automatically detects Lever, Greenhouse, Workday, Ashby, BambooHR, and many more job platforms.

---

### Option 2: Bookmarklet (Alternative)

**Note:** Requires manual activation each time. Userscript is recommended for better UX.

#### 1. Trust the Certificate (One-time)

On your phone/device:
1. Open browser and visit `https://YOUR_SERVER_IP:PORT/script.js`
2. Accept security warning about the self-signed certificate
3. You should see JavaScript code displayed
4. Your device now trusts the certificate

#### 2. Create the Bookmarklet

Create a bookmark with this JavaScript as the URL:

```javascript
javascript:(function(){var s=document.createElement('script');s.src='https://YOUR_SERVER_IP:PORT/script.js';document.body.appendChild(s);})();
```

Replace `YOUR_SERVER_IP:PORT` with your server address.

#### 3. Use It

1. Navigate to any job application form
2. Open bookmarks and tap the bookmarklet
3. The autofill menu appears

## How It Works

### Userscript Mode:
1. Userscript auto-detects job application pages
2. Injects `script.js` automatically when forms are detected
3. Floating ✨ button appears in bottom-right corner
4. Click → Fill Constants → Type "##" in complex fields → Fill LLM Fields

### Workflow:
1. **Fill Constants** - Instantly fills name, email, phone, etc. from `constants.json`
2. **Type "##"** - Mark any empty field where you want AI to write a custom answer
3. **Fill LLM Fields** - AI generates personalized responses for marked fields
4. **Done!** - Review and submit

The system uses:
- **Exact matching** for known fields (instant)
- **Semantic matching** with embeddings for similar fields (fast)
- **LLM fallback** only for complex/unusual questions (slower but smart)

## Security Notes

- Uses self-signed SSL certificate (you'll see a warning on first visit)
- HTTPS is required to work on modern HTTPS websites (mixed content blocking)
- Server is exposed on your public IP - consider firewall rules if needed
