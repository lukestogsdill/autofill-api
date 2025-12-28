// ==UserScript==
// @name         Autofill Assistant
// @namespace    https://lustogs.com
// @version      3.0
// @description  Just a button - click to load, then click menu options to call API
// @author       Luke Stogsdill
// @match        *://*/*
// @grant        none
// @run-at       document-end
// @updateURL    https://YOUR_SERVER_IP:PORT/autofill.user.js
// @downloadURL  https://YOUR_SERVER_IP:PORT/autofill.user.js
// ==/UserScript==

(function() {
    'use strict';

    const API_URL = 'https://YOUR_SERVER_IP:PORT/api/fill';
    let scriptLoaded = false;

    function hasFormOnPage() {
        return document.querySelectorAll('form input, form textarea, form select').length > 0;
    }

    function createButton() {
        if (document.getElementById('autofill-toggle-btn') || !hasFormOnPage()) {
            return;
        }

        const btn = document.createElement('button');
        btn.id = 'autofill-toggle-btn';
        btn.innerHTML = '✨';
        btn.title = 'Autofill';

        Object.assign(btn.style, {
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            width: '56px',
            height: '56px',
            borderRadius: '50%',
            background: 'linear-gradient(135deg, #3b82f6, #2563eb)',
            border: '2px solid rgba(255,255,255,0.3)',
            color: 'white',
            fontSize: '24px',
            cursor: 'pointer',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
            zIndex: '999998',
            transition: 'transform 0.2s'
        });

        btn.onmouseenter = () => btn.style.transform = 'scale(1.1)';
        btn.onmouseleave = () => btn.style.transform = 'scale(1)';

        btn.onclick = () => {
            if (!scriptLoaded) {
                btn.innerHTML = '⏳';
                const script = document.createElement('script');
                script.src = API_URL.replace('/api/fill', '/script.js');
                script.onload = () => {
                    scriptLoaded = true;
                    btn.remove();
                };
                script.onerror = () => {
                    btn.innerHTML = '❌';
                    setTimeout(() => btn.innerHTML = '✨', 2000);
                };
                document.body.appendChild(script);
            }
        };

        document.body.appendChild(btn);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', createButton);
    } else {
        createButton();
    }

    new MutationObserver(() => createButton()).observe(document.body, {
        childList: true,
        subtree: true
    });

})();
