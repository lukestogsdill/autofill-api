(function() {
  'use strict';

  // ============================================================================
  // CONFIGURATION
  // ============================================================================

  const CONFIG = {
    apiBase: API_URL.replace('/api/fill', ''),
    llmTrigger: 'llm-fill',  // Type this in a form field to trigger LLM for that field
    debugMode: true
  };

  // ============================================================================
  // STYLES
  // ============================================================================

  const STYLES = {
    modal: `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0,0,0,0.85);
      backdrop-filter: blur(8px);
      z-index: 999999;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    `,
    content: `
      background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
      border-radius: 16px;
      padding: 0;
      width: 420px;
      max-width: 90%;
      max-height: 85%;
      overflow: hidden;
      box-shadow: 0 20px 60px rgba(0,0,0,0.5);
    `,
    header: `
      padding: 20px 24px;
      border-bottom: 1px solid rgba(255,255,255,0.1);
      display: flex;
      justify-content: space-between;
      align-items: center;
    `,
    title: `
      margin: 0;
      font-size: 24px;
      font-weight: 700;
      color: white;
    `,
    closeBtn: `
      width: 32px;
      height: 32px;
      border-radius: 8px;
      background: rgba(255,255,255,0.1);
      border: 1px solid rgba(255,255,255,0.2);
      color: white;
      font-size: 18px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
    `,
    settingsBtn: `
      width: 32px;
      height: 32px;
      border-radius: 8px;
      background: rgba(255,255,255,0.1);
      border: 1px solid rgba(255,255,255,0.2);
      color: white;
      font-size: 18px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
    `,
    body: `
      padding: 24px;
      max-height: 500px;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
    `,
    button: `
      width: 100%;
      padding: 18px 20px;
      margin-bottom: 12px;
      font-size: 16px;
      font-weight: 600;
      border: none;
      border-radius: 12px;
      cursor: pointer;
      transition: all 0.2s;
      color: white;
      text-align: left;
      display: flex;
      align-items: center;
      justify-content: space-between;
    `,
    buttonPrimary: `
      background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
      border: 1px solid rgba(255,255,255,0.3);
    `,
    buttonSuccess: `
      background: linear-gradient(135deg, #10b981 0%, #059669 100%);
      border: 1px solid rgba(255,255,255,0.3);
    `,
    buttonWarning: `
      background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
      border: 1px solid rgba(255,255,255,0.3);
    `,
    infoBox: `
      background: rgba(255,255,255,0.08);
      padding: 14px 16px;
      border-radius: 10px;
      margin-bottom: 12px;
      border: 1px solid rgba(255,255,255,0.15);
    `,
    infoLabel: `
      font-size: 12px;
      color: rgba(255,255,255,0.5);
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 4px;
    `,
    fieldInput: `
      width: 100%;
      padding: 10px 12px;
      margin-top: 6px;
      border: 1px solid rgba(255,255,255,0.2);
      border-radius: 8px;
      background: rgba(255,255,255,0.1);
      color: white;
      font-size: 14px;
      font-family: inherit;
    `
  };

  // ============================================================================
  // DEBUG LOGGER
  // ============================================================================

  const Logger = {
    logs: [],

    log(...args) {
      console.log(...args);
      if (CONFIG.debugMode) {
        const logEntry = args.map(arg =>
          typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
        ).join(' ');
        this.logs.push(logEntry);
      }
    },

    show() {
      // Deprecated - logs are now shown in the logs menu
      // Keeping this method for backwards compatibility
    },

    clear() {
      this.logs = [];
    }
  };

  // ============================================================================
  // HTML TEMPLATES
  // ============================================================================

  const Templates = {
    mainMenu() {
      return `
        <div id="autofill-modal" style="${STYLES.modal}">
          <div style="${STYLES.content}">
            <div style="${STYLES.header}">
              <h2 style="${STYLES.title}">Autofill</h2>
              <div style="display: flex; gap: 8px;">
                <button id="logs-btn" style="${STYLES.settingsBtn}">📋</button>
                <button id="settings-btn" style="${STYLES.settingsBtn}">⚙️</button>
                <button id="close-btn" style="${STYLES.closeBtn}">✕</button>
              </div>
            </div>
            <div style="${STYLES.body}">
              <div style="background: rgba(59,130,246,0.1); border: 1px solid rgba(59,130,246,0.3); border-radius: 8px; padding: 12px; margin-bottom: 16px;">
                <div style="font-size: 12px; color: rgba(255,255,255,0.7); line-height: 1.4;">
                  <strong style="color: #60a5fa;">Workflow:</strong><br/>
                  1. Fill Constants → 2. Type "##" in empty fields → 3. Fill LLM Fields → 4. Save to Doc
                </div>
              </div>
              <button id="fill-constants-btn" style="${STYLES.button}${STYLES.buttonPrimary}">
                <span>1. Fill Constants</span>
                <span style="font-size: 20px;">📝</span>
              </button>
              <button id="fill-llm-btn" style="${STYLES.button}${STYLES.buttonSuccess}">
                <span>2. Fill LLM Fields</span>
                <span style="font-size: 20px;">⚡</span>
              </button>
              <button id="save-doc-btn" style="${STYLES.button}${STYLES.buttonWarning}">
                <span>3. Save to Google Doc</span>
                <span style="font-size: 20px;">📄</span>
              </button>
            </div>
          </div>
        </div>
      `;
    },

    settingsMenu(constants) {
      const constantsHTML = Object.entries(constants).map(([key, value]) => `
        <div style="${STYLES.infoBox} display: flex; gap: 8px; align-items: start;">
          <div style="flex: 1;">
            <div style="${STYLES.infoLabel}">${key.replace(/_/g, ' ').toUpperCase()}</div>
            <input
              type="text"
              data-key="${key}"
              value="${value}"
              placeholder="constant value"
              style="${STYLES.fieldInput}"
            />
          </div>
          <button
            data-remove="${key}"
            style="
              width: 32px;
              height: 32px;
              border-radius: 8px;
              background: rgba(239, 68, 68, 0.2);
              border: 1px solid rgba(239, 68, 68, 0.3);
              color: #ef4444;
              cursor: pointer;
              font-size: 18px;
              margin-top: 20px;
            "
          >✕</button>
        </div>
      `).join('');

      return `
        <div id="autofill-modal" style="${STYLES.modal}">
          <div style="${STYLES.content}">
            <div style="${STYLES.header}">
              <h2 style="${STYLES.title}">Constants</h2>
              <button id="close-btn" style="${STYLES.closeBtn}">✕</button>
            </div>
            <div style="padding: 0; overflow: hidden; display: flex; flex-direction: column; height: 500px;">
              <div style="padding: 16px 24px; border-bottom: 1px solid rgba(255,255,255,0.1); display: flex; gap: 12px;">
                <button id="save-constants-btn" style="${STYLES.button}${STYLES.buttonPrimary} flex: 1; margin-bottom: 0; padding: 12px 16px;">
                  <span>Save</span><span>💾</span>
                </button>
                <button id="add-field-btn" style="${STYLES.button}${STYLES.buttonSuccess} flex: 1; margin-bottom: 0; padding: 12px 16px;">
                  <span>Add Field</span><span>➕</span>
                </button>
              </div>
              <div id="constants-scroll" style="flex: 1; overflow-y: auto; padding: 24px;">
                ${constantsHTML}
              </div>
            </div>
          </div>
        </div>
      `;
    },

    newConstantField() {
      return `
        <div style="${STYLES.infoBox} display: flex; gap: 8px; align-items: start;">
          <div style="flex: 1;">
            <input
              type="text"
              data-new-key
              placeholder="field_name"
              style="${STYLES.fieldInput} margin-top: 0; font-family: monospace; font-size: 12px;"
            />
            <input
              type="text"
              data-new-value
              placeholder="value"
              style="${STYLES.fieldInput}"
            />
          </div>
          <button
            data-remove
            style="
              width: 32px;
              height: 32px;
              border-radius: 8px;
              background: rgba(239, 68, 68, 0.2);
              border: 1px solid rgba(239, 68, 68, 0.3);
              color: #ef4444;
              cursor: pointer;
              font-size: 18px;
              margin-top: 0;
            "
          >✕</button>
        </div>
      `;
    },

    logsView(logs) {
      const logsHTML = logs.length > 0
        ? logs.slice().reverse().map(log => `<div style="margin-bottom: 8px; padding: 8px; background: rgba(0,0,0,0.3); border-radius: 6px; font-size: 12px; font-family: monospace; word-wrap: break-word;">${log}</div>`).join('')
        : '<div style="color: rgba(255,255,255,0.5); text-align: center; padding: 20px;">No logs yet</div>';

      return `
        <div id="autofill-modal" style="${STYLES.modal}">
          <div style="${STYLES.content}">
            <div style="${STYLES.header}">
              <h2 style="${STYLES.title}">Debug Logs</h2>
              <div style="display: flex; gap: 8px;">
                <button id="clear-logs-btn" style="${STYLES.settingsBtn}">🗑️</button>
                <button id="close-btn" style="${STYLES.closeBtn}">✕</button>
              </div>
            </div>
            <div style="padding: 24px; max-height: 500px; overflow-y: auto; -webkit-overflow-scrolling: touch;">
              ${logsHTML}
            </div>
          </div>
        </div>
      `;
    }
  };

  // ============================================================================
  // API SERVICE
  // ============================================================================

  const API = {
    async getConstants() {
      const response = await fetch(`${CONFIG.apiBase}/api/constants`);
      if (!response.ok) throw new Error('Failed to fetch constants');
      return await response.json();
    },

    async saveConstants(constants) {
      const response = await fetch(`${CONFIG.apiBase}/api/constants`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(constants)
      });
      if (!response.ok) throw new Error('Failed to save constants');
      return await response.json();
    },

    async getJobContext() {
      try {
        const response = await fetch(`${CONFIG.apiBase}/api/context`);
        if (!response.ok) throw new Error('Failed to fetch job context');
        return await response.json();
      } catch (error) {
        Logger.log('Error fetching job context:', error);
        return { title: 'Unknown', company: 'Unknown', url: window.location.href };
      }
    },

    async fillConstants(fields) {
      const response = await fetch(`${CONFIG.apiBase}/api/fill-constants`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ fields })
      });
      if (!response.ok) throw new Error(`API error: ${response.status}`);
      return await response.json();
    },

    async fillLLM(fields, context) {
      const response = await fetch(`${CONFIG.apiBase}/api/fill-llm`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ fields, job_context: context })
      });
      if (!response.ok) throw new Error(`API error: ${response.status}`);
      return await response.json();
    },

    async saveToGoogleDoc(formData) {
      const response = await fetch(`${CONFIG.apiBase}/api/save-doc`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });
      if (!response.ok) throw new Error(`Failed to save to Google Doc: ${response.status}`);
      return await response.json();
    }
  };

  // ============================================================================
  // FORM FIELD HANDLER
  // ============================================================================

  const FormHandler = {
    collectFields() {
      Logger.log('🔍 Collecting form fields...');
      const fields = [];
      const elements = new Map();
      const forms = document.querySelectorAll('form');

      Logger.log(`Found ${forms.length} forms`);

      if (!forms.length) {
        alert('No forms found on this page!');
        return null;
      }

      forms.forEach(form => {
        const inputs = form.querySelectorAll('input, textarea, select');
        inputs.forEach(input => {
          // Skip buttons and hidden fields
          if (['submit', 'button', 'hidden'].includes(input.type)) return;

          const fieldId = this.generateFieldId(input);
          const label = this.getFieldLabel(input);

          const field = {
            id: fieldId,
            name: input.name || '',
            type: input.type || input.tagName.toLowerCase(),
            label: label,
            placeholder: input.placeholder || '',
            required: input.required || false,
            value: input.value || ''
          };

          // Handle select options
          if (input.tagName === 'SELECT') {
            field.options = Array.from(input.options).map(o => ({
              value: o.value,
              text: o.text
            }));
          }

          // Handle radio buttons (group them)
          if (input.type === 'radio') {
            const existingGroup = fields.find(f => f.name === input.name && f.type === 'radio');
            if (existingGroup) {
              existingGroup.options.push({
                value: input.value,
                text: this.getFieldLabel(input) || input.value
              });
              existingGroup.elements.push(input);
            } else {
              field.options = [{
                value: input.value,
                text: this.getFieldLabel(input) || input.value
              }];
              field.elements = [input];
              fields.push(field);
              elements.set(fieldId, input);
            }
            return;
          }

          fields.push(field);
          elements.set(fieldId, input);
        });
      });

      Logger.log(`✅ Collected ${fields.length} fields`);
      return { fields, elements };
    },

    generateFieldId(input) {
      if (input.name) return input.name;
      if (input.id) return input.id;

      const label = this.getFieldLabel(input);
      return label.toLowerCase().replace(/[^a-z0-9]/g, '_') ||
             `field_${Math.random().toString(36).substr(2, 9)}`;
    },

    getFieldLabel(input) {
      // Try label[for="id"]
      if (input.id) {
        const label = document.querySelector(`label[for="${input.id}"]`);
        if (label) return label.textContent.trim();
      }

      // Try parent label
      const parentLabel = input.closest('label');
      if (parentLabel) {
        return parentLabel.textContent.replace(input.value, '').trim();
      }

      // Try previous sibling
      const prevText = input.previousElementSibling;
      if (prevText && prevText.textContent) {
        return prevText.textContent.trim();
      }

      // Try aria-label
      if (input.getAttribute('aria-label')) {
        return input.getAttribute('aria-label');
      }

      return input.placeholder || input.name || '';
    },

    getEmptyFields(fields) {
      return fields.filter(field => {
        if (field.type === 'checkbox' || field.type === 'radio') {
          return true; // Always include
        }
        return !field.value || field.value.trim() === '';
      });
    },

    /**
     * Get fields that have been marked with "##" by the user
     * Re-scans the page to get current values
     */
    getLLMMarkedFields() {
      Logger.log('🔍 Scanning for fields marked with "##"...');
      const llmFields = [];
      const elements = new Map();
      const forms = document.querySelectorAll('form');

      forms.forEach(form => {
        const inputs = form.querySelectorAll('input, textarea, select');
        inputs.forEach(input => {
          // Skip buttons and hidden fields
          if (['submit', 'button', 'hidden'].includes(input.type)) return;

          // Check if field value is "##"
          if (input.value && input.value.trim() === CONFIG.llmTrigger) {
            const fieldId = this.generateFieldId(input);
            const label = this.getFieldLabel(input);

            const field = {
              id: fieldId,
              name: input.name || '',
              type: input.type || input.tagName.toLowerCase(),
              label: label,
              placeholder: input.placeholder || '',
              required: input.required || false,
              value: ''  // Clear the ## placeholder
            };

            // Handle select options
            if (input.tagName === 'SELECT') {
              field.options = Array.from(input.options).map(o => ({
                value: o.value,
                text: o.text
              }));
            }

            llmFields.push(field);
            elements.set(fieldId, input);
            Logger.log(`✓ Found LLM field: "${label}" (${fieldId})`);
          }
        });
      });

      Logger.log(`✅ Found ${llmFields.length} fields marked for LLM`);
      return { fields: llmFields, elements };
    }
  };

  // ============================================================================
  // FUZZY MATCHER
  // ============================================================================

  const FuzzyMatcher = {
    /**
     * Match form fields to constants using fuzzy matching on labels, names, placeholders
     */
    matchFieldsToConstants(fields, constants) {
      Logger.log('🔗 Fuzzy matching fields to constants...');
      const matches = {};

      fields.forEach(field => {
        const searchTerms = [
          field.label,
          field.name,
          field.placeholder,
          field.id
        ].filter(Boolean).map(s => s.toLowerCase());

        // Try to find best matching constant
        let bestMatch = null;
        let bestScore = 0;

        Object.entries(constants).forEach(([key, value]) => {
          const constantKey = key.toLowerCase().replace(/_/g, ' ');

          searchTerms.forEach(term => {
            const score = this.similarityScore(term, constantKey);
            if (score > bestScore && score > 0.5) { // threshold
              bestScore = score;
              bestMatch = { key, value };
            }
          });
        });

        if (bestMatch) {
          matches[field.id] = bestMatch.value;
          Logger.log(`✓ Matched "${field.id}" to "${bestMatch.key}" (score: ${bestScore.toFixed(2)})`);
        }
      });

      return matches;
    },

    /**
     * Simple similarity score (0-1) based on common words and substring matching
     */
    similarityScore(str1, str2) {
      str1 = str1.toLowerCase();
      str2 = str2.toLowerCase();

      // Exact match
      if (str1 === str2) return 1.0;

      // Substring match
      if (str1.includes(str2) || str2.includes(str1)) return 0.9;

      // Common words
      const words1 = str1.split(/\s+/);
      const words2 = str2.split(/\s+/);
      const commonWords = words1.filter(w => words2.includes(w));

      if (commonWords.length > 0) {
        return 0.7 * (commonWords.length / Math.max(words1.length, words2.length));
      }

      // Levenshtein-like: count common characters
      const common = [...str1].filter(c => str2.includes(c)).length;
      return 0.5 * (common / Math.max(str1.length, str2.length));
    }
  };

  // ============================================================================
  // FORM FILLER
  // ============================================================================

  const FormFiller = {
    fill(fieldElements, data, collectedFields) {
      Logger.log('📝 Filling form with data...');
      let filledCount = 0;

      fieldElements.forEach((element, fieldId) => {
        let value = data[fieldId];

        // Special handling: password fields
        if (element.type === 'password') {
          const passwordKey = Object.keys(data).find(key =>
            key.toLowerCase() === 'password'
          );
          if (passwordKey) {
            value = data[passwordKey];
            Logger.log(`🔑 Password field "${fieldId}" filled`);
          }
        }

        if (!value) {
          Logger.log(`⊘ Skipping "${fieldId}" - no value`);
          return;
        }

        // Apply value based on field type
        if (element.tagName === 'SELECT') {
          for (let option of element.options) {
            if (option.value === value ||
                option.text.toLowerCase().includes(value.toLowerCase())) {
              element.value = option.value;
              filledCount++;
              Logger.log(`✓ Filled SELECT "${fieldId}"`);
              break;
            }
          }
        } else if (element.type === 'checkbox') {
          element.checked = value === true || value === 'true' || value === 'yes';
          filledCount++;
          Logger.log(`✓ Filled CHECKBOX "${fieldId}"`);
        } else if (element.type === 'radio') {
          const field = collectedFields.find(f => f.id === fieldId);
          if (field && field.elements) {
            field.elements.forEach(radio => {
              if (radio.value === value ||
                  radio.value.toLowerCase().includes(value.toLowerCase())) {
                radio.checked = true;
                filledCount++;
                Logger.log(`✓ Filled RADIO "${fieldId}"`);
              }
            });
          }
        } else {
          element.value = value;
          filledCount++;
          Logger.log(`✓ Filled "${fieldId}"`);
        }

        // Trigger events for frameworks
        element.dispatchEvent(new Event('input', { bubbles: true }));
        element.dispatchEvent(new Event('change', { bubbles: true }));
      });

      Logger.log(`✅ Filled ${filledCount} fields total`);
      return filledCount;
    }
  };

  // ============================================================================
  // UI CONTROLLER
  // ============================================================================

  const UI = {
    state: {
      collectedFields: [],
      fieldElements: new Map()
    },

    showMainMenu() {
      this.closeModal();
      document.body.insertAdjacentHTML('beforeend', Templates.mainMenu());

      // Attach event listeners
      document.getElementById('close-btn').onclick = () => this.closeModal();
      document.getElementById('logs-btn').onclick = () => this.showLogs();
      document.getElementById('settings-btn').onclick = () => this.showSettings();
      document.getElementById('fill-constants-btn').onclick = () => this.fillWithConstants();
      document.getElementById('fill-llm-btn').onclick = () => this.fillLLMFields();
      document.getElementById('save-doc-btn').onclick = () => this.saveToDoc();

      // Close on backdrop click
      document.getElementById('autofill-modal').onclick = (e) => {
        if (e.target.id === 'autofill-modal') this.closeModal();
      };
    },

    showLogs() {
      this.closeModal();
      document.body.insertAdjacentHTML('beforeend', Templates.logsView(Logger.logs));

      // Attach event listeners
      document.getElementById('close-btn').onclick = () => this.showMainMenu();
      document.getElementById('clear-logs-btn').onclick = () => {
        Logger.clear();
        this.showLogs(); // Refresh
      };
    },

    async showSettings() {
      try {
        const constants = await API.getConstants();
        this.closeModal();
        document.body.insertAdjacentHTML('beforeend', Templates.settingsMenu(constants));

        // Attach event listeners
        document.getElementById('close-btn').onclick = () => this.showMainMenu();
        document.getElementById('save-constants-btn').onclick = () => this.saveConstants();
        document.getElementById('add-field-btn').onclick = () => this.addNewConstantField();

        // Remove buttons
        document.querySelectorAll('[data-remove]').forEach(btn => {
          btn.onclick = () => btn.closest('[style*="infoBox"]').remove();
        });
      } catch (error) {
        alert(`Error loading settings: ${error.message}`);
      }
    },

    addNewConstantField() {
      const container = document.getElementById('constants-scroll');
      const tempDiv = document.createElement('div');
      tempDiv.innerHTML = Templates.newConstantField();
      const newField = tempDiv.firstElementChild;

      newField.querySelector('[data-remove]').onclick = () => newField.remove();
      container.appendChild(newField);
      container.scrollTop = container.scrollHeight;
    },

    async saveConstants() {
      const saveBtn = document.getElementById('save-constants-btn');
      const container = document.getElementById('constants-scroll');

      try {
        saveBtn.innerHTML = '<span>Saving...</span>';
        saveBtn.disabled = true;

        const updated = {};

        // Existing constants
        container.querySelectorAll('input[data-key]').forEach(input => {
          const key = input.getAttribute('data-key');
          if (key) updated[key] = input.value;
        });

        // New constants
        const newKeys = container.querySelectorAll('input[data-new-key]');
        const newValues = container.querySelectorAll('input[data-new-value]');
        newKeys.forEach((keyInput, idx) => {
          const key = keyInput.value.trim();
          const value = newValues[idx].value.trim();
          if (key && value) updated[key] = value;
        });

        await API.saveConstants(updated);

        saveBtn.innerHTML = '<span>✓ Saved!</span>';
        saveBtn.style.background = 'linear-gradient(135deg, #10b981 0%, #059669 100%)';

        setTimeout(() => this.showMainMenu(), 1000);
      } catch (error) {
        alert(`Error saving: ${error.message}`);
        saveBtn.innerHTML = '<span>Save</span><span>💾</span>';
        saveBtn.disabled = false;
      }
    },

    async fillWithConstants() {
      Logger.clear();

      try {
        const result = FormHandler.collectFields();
        if (!result) return;

        this.state.collectedFields = result.fields;
        this.state.fieldElements = result.elements;

        const emptyFields = FormHandler.getEmptyFields(result.fields);

        if (emptyFields.length === 0) {
          alert('All fields are already filled!');
          this.closeModal();
          return;
        }

        // Send empty fields to backend for constants matching
        const cleanFields = emptyFields.map(field => ({
          id: field.id,
          name: field.name,
          type: field.type,
          label: field.label,
          placeholder: field.placeholder,
          required: field.required,
          value: field.value,
          options: field.options
        }));

        const data = await API.fillConstants(cleanFields);
        const matches = data.fields || {};

        const filledCount = FormFiller.fill(
          this.state.fieldElements,
          matches,
          this.state.collectedFields
        );

        alert(`✓ Filled ${filledCount}/${emptyFields.length} fields with constants!`);
        this.closeModal();
      } catch (error) {
        alert(`Error: ${error.message}`);
      } finally {
        if (CONFIG.debugMode) Logger.show();
      }
    },

    async fillLLMFields() {
      Logger.clear();

      try {
        // Scan page for fields marked with "##"
        const llmResult = FormHandler.getLLMMarkedFields();

        if (!llmResult || llmResult.fields.length === 0) {
          alert('No fields marked with "##" found!\n\nType "##" in any field you want AI to complete.');
          return;
        }

        Logger.log(`Found ${llmResult.fields.length} fields marked for LLM`);

        // Store elements for later filling
        this.state.collectedFields = llmResult.fields;
        this.state.fieldElements = llmResult.elements;

        // Get job context
        const context = await API.getJobContext();

        // Send marked fields to LLM
        const cleanFields = llmResult.fields.map(field => ({
          id: field.id,
          name: field.name,
          type: field.type,
          label: field.label,
          placeholder: field.placeholder,
          required: field.required,
          value: field.value,
          options: field.options
        }));

        const llmData = await API.fillLLM(cleanFields, context);
        const llmMatches = llmData.fields || {};

        const filledCount = FormFiller.fill(
          this.state.fieldElements,
          llmMatches,
          this.state.collectedFields
        );

        alert(`✓ Filled ${filledCount}/${llmResult.fields.length} fields with LLM!`);
        this.closeModal();
      } catch (error) {
        alert(`Error: ${error.message}`);
      } finally {
        if (CONFIG.debugMode) Logger.show();
      }
    },

    async saveToDoc() {
      try {
        const result = FormHandler.collectFields();
        if (!result) return;

        // Collect all filled form data
        const formData = {};
        result.elements.forEach((element, fieldId) => {
          if (element.value && element.value.trim() !== '') {
            formData[fieldId] = element.value;
          }
        });

        if (Object.keys(formData).length === 0) {
          alert('No data to save! Fill out the form first.');
          return;
        }

        // TODO: Implement Google Doc save functionality
        alert('Google Doc save coming soon!\n\nCollected data:\n' + JSON.stringify(formData, null, 2));

        this.closeModal();
      } catch (error) {
        alert(`Error: ${error.message}`);
      }
    },

    closeModal() {
      const modal = document.getElementById('autofill-modal');
      if (modal) modal.remove();
    }
  };

  // ============================================================================
  // FLOATING TOGGLE BUTTON
  // ============================================================================

  const FloatingButton = {
    button: null,

    create() {
      // Check if button already exists
      if (document.getElementById('autofill-toggle-btn')) {
        Logger.log('Floating button already exists');
        return;
      }

      this.button = document.createElement('button');
      this.button.id = 'autofill-toggle-btn';
      this.button.innerHTML = '✨';
      this.button.title = 'Toggle Autofill Menu';

      // Styles
      Object.assign(this.button.style, {
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        width: '56px',
        height: '56px',
        borderRadius: '50%',
        background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
        border: '2px solid rgba(255,255,255,0.3)',
        color: 'white',
        fontSize: '24px',
        cursor: 'pointer',
        boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
        zIndex: '999998',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        transition: 'all 0.2s',
        justifyContent: 'center',
        transition: 'all 0.2s',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
      });

      // Hover effect
      this.button.onmouseenter = () => {
        this.button.style.transform = 'scale(1.1)';
        this.button.style.boxShadow = '0 6px 16px rgba(0,0,0,0.4)';
      };
      this.button.onmouseleave = () => {
        this.button.style.transform = 'scale(1)';
        this.button.style.boxShadow = '0 4px 12px rgba(0,0,0,0.3)';
      };

      // Click handler
      this.button.onclick = () => {
        const modal = document.getElementById('autofill-modal');
        if (modal) {
          UI.closeModal();
        } else {
          UI.showMainMenu();
        }
      };

      document.body.appendChild(this.button);
      Logger.log('✨ Floating toggle button created');
    },

    remove() {
      if (this.button) {
        this.button.remove();
        this.button = null;
      }
    }
  };

  // ============================================================================
  // INITIALIZE
  // ============================================================================

  // Create floating button (only once)
  FloatingButton.create();

  // Show menu on first load
  UI.showMainMenu();

})();
