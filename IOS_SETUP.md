# iOS Safari Userscript Setup Guide

## Step 1: Install Userscripts Extension

1. Open **App Store** on your iPhone
2. Search for **"Userscripts"** (by Justin Wasack)
3. Install the app (it's free)

## Step 2: Enable Extension in Safari

1. Open **Settings** app on iPhone
2. Scroll down and tap **Safari**
3. Tap **Extensions**
4. Find **Userscripts** and toggle it **ON**
5. Tap **Userscripts** again and set permissions to **Allow**

## Step 3: Trust Your Server Certificate

1. Open **Safari** browser
2. Visit: `https://YOUR_SERVER_IP:PORT/autofill.user.js`
   - Example: `https://168.93.47.2:4444/autofill.user.js`
3. You'll see a security warning about the certificate
4. Tap **Show Details** → **visit this website**
5. You should now see the userscript code (starts with `// ==UserScript==`)

## Step 4: Install the Userscript

**Option A: Using the AA Button (Recommended)**
1. While viewing the userscript file, tap the **AA** button in the address bar
2. Tap **Manage Extensions**
3. Tap **Userscripts**
4. You should see the script listed - tap **Install**

**Option B: Using Userscripts App**
1. Open the **Userscripts** app on your phone
2. Tap the **+** button
3. Paste the URL: `https://YOUR_SERVER_IP:PORT/autofill.user.js`
4. Tap **Add** or **Install**

**Option C: Manual Installation**
1. Copy all the userscript code from Safari
2. Open the **Userscripts** app
3. Tap **+** → **New Userscript**
4. Paste the code
5. Save

## Step 5: Test It

1. Visit **any website with a form** in Safari (it now runs on all sites!)
   - Try Google, any job site, or any form page
2. **Look for these signs the userscript is running:**
   - **Blue notification** in top-right saying "Autofill loading..." then "Autofill ready!"
   - **✨ floating button** appears in bottom-right corner
   - **Console logs** (if you have Safari developer mode on)

### How to Check if It's Running:

**Method 1: Visual (Easiest)**
- Visit any page with a form
- Look for the blue notification sliding in from the right
- Look for the ✨ button in the bottom-right corner

**Method 2: Safari Console**
- Settings → Safari → Advanced → Enable "Web Inspector"
- In Safari, tap the **aA** button → **Show Web Inspector**
- Look for logs like:
  ```
  [Autofill Userscript] 🚀 Initialized on: https://...
  [Autofill] Detected X form fields
  [Autofill] ✅ Script loaded successfully!
  ```

**Method 3: Userscripts App**
- Open the Userscripts app
- Check if "Autofill Assistant" is listed and enabled
- You should see version **2.1** (updated)

## Troubleshooting

### "Script not loading on job sites"
- Check that Userscripts extension is **ON** in Safari settings
- Verify the script is installed in the Userscripts app
- Try refreshing the page

### "Can't see the userscript install option"
- Make sure you're visiting `/autofill.user.js` (NOT `/script.js`)
- Verify the extension is enabled in Safari settings
- Try using the Userscripts app directly (Option B above)

### "Certificate error won't go away"
- You need to trust the certificate the first time you visit the URL
- After trusting it once, it will work for all future requests

### "Floating button doesn't appear"
- Open Safari's **Developer Mode**: Settings → Safari → Advanced → Web Inspector
- Check browser console for errors
- Make sure the page has forms on it

## Server URLs

- **Main script**: `https://YOUR_IP:PORT/script.js`
- **Userscript**: `https://YOUR_IP:PORT/autofill.user.js` ← Use this one!
- **API endpoint**: `https://YOUR_IP:PORT/api/fill`

Replace `YOUR_IP:PORT` with your actual server address (e.g., `168.93.47.2:4444`)

## What Sites Are Supported?

**ALL SITES!** The userscript now runs on every website (`@match *://*/*`)

It will:
- ✅ Load on any page with form fields
- ✅ Skip pages without forms automatically
- ✅ Watch for dynamically loaded forms (React/Vue apps)

This means you can use it on:
- Job application sites (Greenhouse, Lever, Workday, etc.)
- Google Forms
- Company career pages
- Any custom application form
- Literally any website with `<input>` fields
