// ==UserScript==
// @name         Autofill Assistant
// @namespace    https://lustogs.com
// @version      2.0
// @description  Automatically loads autofill functionality on job application pages
// @author       Luke Stogsdill
// @match        *://*.lever.co/*
// @match        *://*.greenhouse.io/*
// @match        *://*.myworkdayjobs.com/*
// @match        *://*.taleo.net/*
// @match        *://*.workable.com/*
// @match        *://*.ashbyhq.com/*
// @match        *://*.breezy.hr/*
// @match        *://*.bamboohr.com/*
// @match        *://*.smartrecruiters.com/*
// @match        *://*.jobvite.com/*
// @match        *://*.icims.com/*
// @match        *://*.ultipro.com/*
// @match        *://*.paycom.com/*
// @match        *://*.paylocity.com/*
// @match        *://jobs.*/*
// @match        *://careers.*/*
// @match        *://apply.*/*
// @match        *://recruiting.*/*
// @match        *://*/*/careers/*
// @match        *://*/*/jobs/*
// @match        *://*/*/apply/*
// @match        *://*/*/job/*
// @match        *://*/*/application/*
// @grant        none
// @run-at       document-end
// @updateURL    https://YOUR_SERVER_IP:PORT/autofill.user.js
// @downloadURL  https://YOUR_SERVER_IP:PORT/autofill.user.js
// ==/UserScript==

(function() {
    'use strict';

    // Configuration - REPLACE WITH YOUR SERVER INFO
    const API_URL = 'https://YOUR_SERVER_IP:PORT/api/fill';

    // Check if we're on a page with forms
    function hasFormOnPage() {
        return document.querySelectorAll('form input, form textarea, form select').length > 0;
    }

    // Load the main script
    function loadAutofillScript() {
        // Check if already loaded
        if (document.getElementById('autofill-toggle-btn')) {
            console.log('[Autofill] Script already loaded');
            return;
        }

        // Only load on pages with forms
        if (!hasFormOnPage()) {
            console.log('[Autofill] No forms detected on this page, skipping...');
            return;
        }

        console.log('[Autofill] Loading autofill script...');

        const script = document.createElement('script');
        script.src = API_URL.replace('/api/fill', '/script.js');
        script.onerror = function() {
            console.error('[Autofill] Failed to load script. Check your server connection.');
        };
        script.onload = function() {
            console.log('[Autofill] ✓ Script loaded successfully!');
        };

        document.body.appendChild(script);
    }

    // Wait for page to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', loadAutofillScript);
    } else {
        loadAutofillScript();
    }

    // Also watch for dynamic form loading (e.g., React apps)
    const observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.addedNodes.length > 0) {
                mutation.addedNodes.forEach(function(node) {
                    if (node.nodeType === 1 && (node.tagName === 'FORM' || node.querySelector('form'))) {
                        console.log('[Autofill] New form detected, ensuring script is loaded...');
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
