# Autofill API

A simple HTTPS API server that enables remote form autofilling via a bookmarklet.

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

## Using the Bookmarklet

### 1. Trust the Certificate (One-time setup)

On your phone/device:
1. Open browser and visit `https://YOUR_SERVER_IP:8000/script.js`
2. You'll see a security warning about the self-signed certificate
3. Click "Advanced" → "Proceed anyway" (or similar)
4. You should see the JavaScript code displayed
5. Your device now trusts the certificate

### 2. Create the Bookmarklet

Create a bookmark with this JavaScript as the URL:

```javascript
javascript:(function(){var s=document.createElement('script');s.src='https://YOUR_SERVER_IP:8000/script.js';document.body.appendChild(s);})();
```

Replace `YOUR_SERVER_IP` with your server's public IP.

### 3. Use It

1. Navigate to any form on an HTTPS website
2. Tap the bookmarklet
3. The form will be auto-filled with your data

## How It Works

1. Bookmarklet injects `script.js` into the current page
2. Script collects all form fields and sends them to `/api/fill`
3. Server responds with autofill data
4. Script fills in the form fields

## Security Notes

- Uses self-signed SSL certificate (you'll see a warning on first visit)
- HTTPS is required to work on modern HTTPS websites (mixed content blocking)
- Server is exposed on your public IP - consider firewall rules if needed
