(function(){
  const API_BASE = API_URL.replace('/api/fill', '');

  let collectedFields = [];
  let fieldElements = new Map();

  // Styles
  const styles = {
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

  // Close modal
  function closeModal() {
    const modal = document.getElementById('autofill-modal');
    if (modal) modal.remove();
  }

  // Collect form fields
  function collectFormFields() {
    const fields = [];
    const elements = new Map();

    const forms = document.querySelectorAll('form');
    if (!forms.length) {
      alert('No forms found on this page!');
      return null;
    }

    forms.forEach(form => {
      const inputs = form.querySelectorAll('input, textarea, select');
      inputs.forEach(input => {
        if (input.type === 'submit' || input.type === 'button' || input.type === 'hidden') return;

        const fieldId = input.name || input.id || generateFieldId(input);
        const label = getFieldLabel(input);

        const field = {
          id: fieldId,
          name: input.name || '',
          type: input.type || input.tagName.toLowerCase(),
          label: label,
          placeholder: input.placeholder || '',
          required: input.required || false,
          value: input.value || '',
          element: input
        };

        if (input.tagName === 'SELECT') {
          field.options = Array.from(input.options).map(o => ({
            value: o.value,
            text: o.text
          }));
        }

        if (input.type === 'radio') {
          const radioGroup = fields.find(f => f.name === input.name && f.type === 'radio');
          if (radioGroup) {
            radioGroup.options.push({ value: input.value, text: getFieldLabel(input) || input.value });
            radioGroup.elements.push(input);
          } else {
            field.options = [{ value: input.value, text: getFieldLabel(input) || input.value }];
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

    return { fields, elements };
  }

  function generateFieldId(input) {
    const label = getFieldLabel(input);
    return label.toLowerCase().replace(/[^a-z0-9]/g, '_') || `field_${Math.random().toString(36).substr(2, 9)}`;
  }

  function getFieldLabel(input) {
    if (input.id) {
      const label = document.querySelector(`label[for="${input.id}"]`);
      if (label) return label.textContent.trim();
    }

    const parentLabel = input.closest('label');
    if (parentLabel) {
      return parentLabel.textContent.replace(input.value, '').trim();
    }

    const prevText = input.previousElementSibling;
    if (prevText && prevText.textContent) {
      return prevText.textContent.trim();
    }

    if (input.getAttribute('aria-label')) {
      return input.getAttribute('aria-label');
    }

    return input.placeholder || input.name || '';
  }

  // Fill form with data
  function fillForm(data) {
    let filledCount = 0;

    fieldElements.forEach((element, fieldId) => {
      const value = data[fieldId];
      if (!value) return;

      if (element.tagName === 'SELECT') {
        for (let option of element.options) {
          if (option.value === value || option.text.toLowerCase().includes(value.toLowerCase())) {
            element.value = option.value;
            filledCount++;
            break;
          }
        }
      } else if (element.type === 'checkbox') {
        element.checked = value === true || value === 'true' || value === 'yes';
        filledCount++;
      } else if (element.type === 'radio') {
        const field = collectedFields.find(f => f.id === fieldId);
        if (field && field.elements) {
          field.elements.forEach(radio => {
            if (radio.value === value || radio.value.toLowerCase().includes(value.toLowerCase())) {
              radio.checked = true;
              filledCount++;
            }
          });
        }
      } else {
        element.value = value;
        filledCount++;
      }

      element.dispatchEvent(new Event('input', { bubbles: true }));
      element.dispatchEvent(new Event('change', { bubbles: true }));
    });

    return filledCount;
  }

  // Fetch job context
  async function fetchJobContext() {
    try {
      const response = await fetch(API_BASE + '/api/context');
      if (!response.ok) throw new Error('Failed to fetch job context');
      return await response.json();
    } catch (error) {
      console.error('Error fetching job context:', error);
      return { title: 'Unknown', company: 'Unknown', url: window.location.href };
    }
  }

  // Fill with constants only
  async function fillWithConstants() {
    const result = collectFormFields();
    if (!result) return;

    collectedFields = result.fields;
    fieldElements = result.elements;

    try {
      const context = await fetchJobContext();

      const cleanFields = collectedFields.map(field => ({
        id: field.id,
        name: field.name,
        type: field.type,
        label: field.label,
        placeholder: field.placeholder,
        required: field.required,
        value: field.value,
        options: field.options
      }));

      const response = await fetch(API_BASE + '/api/fill', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fields: cleanFields,
          job_context: context,
          constants_only: true
        })
      });

      if (!response.ok) throw new Error(`API error: ${response.status}`);

      const data = await response.json();
      const filledCount = fillForm(data.fields || data);

      alert(`✓ Filled ${filledCount} fields with constants!`);
      closeModal();
    } catch (error) {
      alert(`Error: ${error.message}`);
    }
  }

  // Fill with LLM
  async function fillWithLLM() {
    const result = collectFormFields();
    if (!result) return;

    collectedFields = result.fields;
    fieldElements = result.elements;

    try {
      const context = await fetchJobContext();

      const cleanFields = collectedFields.map(field => ({
        id: field.id,
        name: field.name,
        type: field.type,
        label: field.label,
        placeholder: field.placeholder,
        required: field.required,
        value: field.value,
        options: field.options
      }));

      const response = await fetch(API_BASE + '/api/fill', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fields: cleanFields,
          job_context: context
        })
      });

      if (!response.ok) throw new Error(`API error: ${response.status}`);

      const data = await response.json();
      const filledCount = fillForm(data.fields || data);

      alert(`✓ Filled ${filledCount} fields with LLM!`);
      closeModal();
    } catch (error) {
      alert(`Error: ${error.message}`);
    }
  }

  // Fill all (constants + LLM)
  async function fillAll() {
    const result = collectFormFields();
    if (!result) return;

    collectedFields = result.fields;
    fieldElements = result.elements;

    try {
      // First try constants
      const constResponse = await fetch(API_BASE + '/api/constants');
      if (constResponse.ok) {
        const constants = await constResponse.json();
        fillForm(constants);
      }

      // Then fill remaining with LLM
      const context = await fetchJobContext();

      const cleanFields = collectedFields.map(field => ({
        id: field.id,
        name: field.name,
        type: field.type,
        label: field.label,
        placeholder: field.placeholder,
        required: field.required,
        value: field.value,
        options: field.options
      }));

      const response = await fetch(API_BASE + '/api/fill', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fields: cleanFields,
          job_context: context
        })
      });

      if (!response.ok) throw new Error(`API error: ${response.status}`);

      const data = await response.json();
      const filledCount = fillForm(data.fields || data);

      alert(`✓ Filled all fields!`);
      closeModal();
    } catch (error) {
      alert(`Error: ${error.message}`);
    }
  }

  // Show settings (manage constants)
  async function showSettings() {
    const existing = document.getElementById('autofill-modal');
    if (existing) existing.remove();

    const modal = document.createElement('div');
    modal.id = 'autofill-modal';
    modal.style.cssText = styles.modal;

    const content = document.createElement('div');
    content.style.cssText = styles.content;

    const header = document.createElement('div');
    header.style.cssText = styles.header;

    const title = document.createElement('h2');
    title.style.cssText = styles.title;
    title.textContent = 'Constants';
    header.appendChild(title);

    const closeBtn = document.createElement('button');
    closeBtn.style.cssText = styles.closeBtn;
    closeBtn.innerHTML = '✕';
    closeBtn.addEventListener('click', showMainMenu);
    header.appendChild(closeBtn);

    const body = document.createElement('div');
    body.style.cssText = 'padding: 0; overflow: hidden; display: flex; flex-direction: column; height: 500px;';
    body.innerHTML = '<div style="text-align: center; color: rgba(255,255,255,0.6); padding: 20px;">Loading...</div>';

    content.appendChild(header);
    content.appendChild(body);
    modal.appendChild(content);
    document.body.appendChild(modal);

    try {
      const response = await fetch(API_BASE + '/api/constants');
      if (!response.ok) throw new Error('Failed to load constants');

      const constants = await response.json();
      body.innerHTML = '';

      // Top bar with Save and Add buttons
      const topBar = document.createElement('div');
      topBar.style.cssText = 'padding: 16px 24px; border-bottom: 1px solid rgba(255,255,255,0.1); display: flex; gap: 12px;';

      const saveBtn = document.createElement('button');
      saveBtn.style.cssText = styles.button + styles.buttonPrimary + 'flex: 1; margin-bottom: 0; padding: 12px 16px;';
      saveBtn.innerHTML = '<span>Save</span><span>💾</span>';
      saveBtn.id = 'save-constants-btn';
      topBar.appendChild(saveBtn);

      const addBtn = document.createElement('button');
      addBtn.style.cssText = styles.button + styles.buttonSuccess + 'flex: 1; margin-bottom: 0; padding: 12px 16px;';
      addBtn.innerHTML = '<span>Add Field</span><span>➕</span>';
      addBtn.addEventListener('click', () => {
        const item = document.createElement('div');
        item.style.cssText = styles.infoBox + 'display: flex; gap: 8px; align-items: start;';
        item.innerHTML = `
          <div style="flex: 1;">
            <input
              type="text"
              data-new-key
              placeholder="field_name"
              style="${styles.fieldInput} margin-top: 0; font-family: monospace; font-size: 12px;"
            />
            <input
              type="text"
              data-new-value
              placeholder="value"
              style="${styles.fieldInput}"
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
        `;
        const removeBtn = item.querySelector('[data-remove]');
        removeBtn.addEventListener('click', () => item.remove());
        scrollContainer.appendChild(item);
        scrollContainer.scrollTop = scrollContainer.scrollHeight;
      });
      topBar.appendChild(addBtn);

      body.appendChild(topBar);

      // Single scroll container
      const scrollContainer = document.createElement('div');
      scrollContainer.style.cssText = 'flex: 1; overflow-y: auto; padding: 24px;';
      scrollContainer.id = 'constants-scroll';

      Object.entries(constants).forEach(([key, value]) => {
        const item = document.createElement('div');
        item.style.cssText = styles.infoBox + 'display: flex; gap: 8px; align-items: start;';
        item.innerHTML = `
          <div style="flex: 1;">
            <div style="${styles.infoLabel}">${key.replace(/_/g, ' ').toUpperCase()}</div>
            <input
              type="text"
              data-key="${key}"
              value="${value}"
              style="${styles.fieldInput}"
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
        `;
        const removeBtn = item.querySelector('[data-remove]');
        removeBtn.addEventListener('click', () => item.remove());
        scrollContainer.appendChild(item);
      });

      body.appendChild(scrollContainer);

      // Save button handler
      saveBtn.addEventListener('click', async () => {
        const inputs = scrollContainer.querySelectorAll('input[data-key]');
        const updated = {};
        inputs.forEach(input => {
          const key = input.getAttribute('data-key');
          if (key) {
            updated[key] = input.value;
          }
        });

        // Add new constants
        const newKeys = scrollContainer.querySelectorAll('input[data-new-key]');
        const newValues = scrollContainer.querySelectorAll('input[data-new-value]');
        newKeys.forEach((keyInput, idx) => {
          const key = keyInput.value.trim();
          const value = newValues[idx].value.trim();
          if (key && value) {
            updated[key] = value;
          }
        });

        try {
          saveBtn.innerHTML = '<span>Saving...</span>';
          saveBtn.disabled = true;

          const saveResponse = await fetch(API_BASE + '/api/constants', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updated)
          });

          if (!saveResponse.ok) throw new Error('Failed to save');

          saveBtn.innerHTML = '<span>✓ Saved!</span>';
          saveBtn.style.background = 'linear-gradient(135deg, #10b981 0%, #059669 100%)';

          setTimeout(() => {
            showMainMenu();
          }, 1000);
        } catch (error) {
          saveBtn.innerHTML = '<span>Error</span>';
          saveBtn.disabled = false;
          setTimeout(() => {
            saveBtn.innerHTML = '<span>Save</span><span>💾</span>';
            saveBtn.style.background = 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)';
          }, 2000);
        }
      });

    } catch (error) {
      body.innerHTML = `<div style="text-align: center; color: #ef4444; padding: 20px;">Error: ${error.message}</div>`;
    }
  }

  // Show main menu
  function showMainMenu() {
    const existing = document.getElementById('autofill-modal');
    if (existing) existing.remove();

    const modal = document.createElement('div');
    modal.id = 'autofill-modal';
    modal.style.cssText = styles.modal;

    const content = document.createElement('div');
    content.style.cssText = styles.content;

    const header = document.createElement('div');
    header.style.cssText = styles.header;

    const title = document.createElement('h2');
    title.style.cssText = styles.title;
    title.textContent = 'Autofill';
    header.appendChild(title);

    const headerRight = document.createElement('div');
    headerRight.style.cssText = 'display: flex; gap: 8px;';

    const settingsBtn = document.createElement('button');
    settingsBtn.style.cssText = styles.settingsBtn;
    settingsBtn.innerHTML = '⚙️';
    settingsBtn.addEventListener('click', showSettings);
    headerRight.appendChild(settingsBtn);

    const closeBtn = document.createElement('button');
    closeBtn.style.cssText = styles.closeBtn;
    closeBtn.innerHTML = '✕';
    closeBtn.addEventListener('click', closeModal);
    headerRight.appendChild(closeBtn);

    header.appendChild(headerRight);

    const body = document.createElement('div');
    body.style.cssText = styles.body;

    const buttons = [
      { text: 'Fill with Constants', icon: '📝', style: styles.buttonPrimary, action: fillWithConstants },
      { text: 'Fill All (LLM)', icon: '⚡', style: styles.buttonSuccess, action: fillAll }
    ];

    buttons.forEach(btn => {
      const button = document.createElement('button');
      button.style.cssText = styles.button + btn.style;
      button.innerHTML = `<span>${btn.text}</span><span style="font-size: 20px;">${btn.icon}</span>`;
      button.addEventListener('click', btn.action);
      body.appendChild(button);
    });

    content.appendChild(header);
    content.appendChild(body);
    modal.appendChild(content);
    document.body.appendChild(modal);

    modal.addEventListener('click', (e) => {
      if (e.target === modal) closeModal();
    });
  }

  // Initialize
  showMainMenu();
})();
