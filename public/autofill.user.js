// ==UserScript==
// @name         Autofill Assistant
// @namespace    https://lustogs.com
// @version      2.1
// @description  Automatically loads autofill functionality on all pages with forms
// @author       Luke Stogsdill
// @match        *://*/*
// @grant        none
// @run-at       document-end
// @updateURL    https://YOUR_SERVER_IP:PORT/autofill.user.js
// @downloadURL  https://YOUR_SERVER_IP:PORT/autofill.user.js
// ==/UserScript==

(function() {
    'use strict';

    // Configuration - REPLACE WITH YOUR SERVER INFO
    const API_URL = 'https://YOUR_SERVER_IP:PORT/api/fill';

    console.log('[Autofill Userscript] 🚀 Initialized on:', window.location.href);

    // Check if we're on a page with forms
    function hasFormOnPage() {
        const forms = document.querySelectorAll('form input, form textarea, form select');
        console.log('[Autofill] Detected', forms.length, 'form fields');
        return forms.length > 0;
    }

    // Show visual notification
    function showNotification(message, emoji = '✨') {
        const notification = document.createElement('div');
        notification.textContent = `${emoji} ${message}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
            color: white;
            padding: 12px 20px;
            border-radius: 8px;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            font-size: 14px;
            font-weight: 600;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            z-index: 999999;
            animation: slideIn 0.3s ease-out;
        `;

        const style = document.createElement('style');
        style.textContent = `
            @keyframes slideIn {
                from { transform: translateX(400px); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
        `;
        document.head.appendChild(style);
        document.body.appendChild(notification);

        setTimeout(() => {
            notification.style.transition = 'all 0.3s';
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(400px)';
            setTimeout(() => notification.remove(), 300);
        }, 2000);
    }

    // Load the main script
    function loadAutofillScript() {
        // Check if already loaded
        if (document.getElementById('autofill-toggle-btn')) {
            console.log('[Autofill] Script already loaded (button exists)');
            return;
        }

        // Only load on pages with forms
        if (!hasFormOnPage()) {
            console.log('[Autofill] No forms detected, skipping...');
            return;
        }

        console.log('[Autofill] 📝 Forms detected! Loading autofill script...');
        showNotification('Autofill loading...', '📝');

        const script = document.createElement('script');
        script.src = API_URL.replace('/api/fill', '/script.js');
        script.onerror = function() {
            console.error('[Autofill] ❌ Failed to load script. Check server connection.');
            showNotification('Autofill failed to load', '❌');
        };
        script.onload = function() {
            console.log('[Autofill] ✅ Script loaded successfully! Look for the ✨ button.');
            showNotification('Autofill ready!', '✅');
        };

        document.body.appendChild(script);
    }

    // Wait for page to be ready
    if (document.readyState === 'loading') {
        console.log('[Autofill] Waiting for page to load...');
        document.addEventListener('DOMContentLoaded', loadAutofillScript);
    } else {
        console.log('[Autofill] Page already loaded, checking for forms...');
        loadAutofillScript();
    }

    // Also watch for dynamic form loading (e.g., React apps)
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.addedNodes.length > 0) {
                mutation.addedNodes.forEach(function(node) {
                    if (node.nodeType === 1 && (node.tagName === 'FORM' || node.querySelector('form'))) {
                        console.log('[Autofill] 🔄 New form detected, loading script...');
                        loadAutofillScript();
                    }
                });
            }
        });
    });

    observer.observe(document.body, {
        childList: true,
        subtree: true
    });

})();
