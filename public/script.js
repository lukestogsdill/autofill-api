(function(){
  // API_URL will be injected by the server
  
  const forms = document.querySelectorAll('form');
  if(!forms.length) { 
    alert('No forms found!'); 
    return; 
  }
  
  const fields = [];
  forms.forEach(form => {
    const inputs = form.querySelectorAll('input, textarea, select');
    inputs.forEach(input => {
      if(input.type === 'submit' || input.type === 'button') return;
      
      const field = {
        name: input.name || input.id || '',
        type: input.type || input.tagName.toLowerCase(),
        label: getLabel(input),
        value: input.value || '',
        required: input.required || false,
        placeholder: input.placeholder || ''
      };
      
      if(input.tagName === 'SELECT') {
        field.options = Array.from(input.options).map(o => o.text);
      }
      
      fields.push(field);
    });
  });
  
  function getLabel(input) {
    const label = document.querySelector(`label[for="${input.id}"]`);
    if(label) return label.textContent.trim();
    
    const parent = input.closest('label');
    if(parent) return parent.textContent.replace(input.value, '').trim();
    
    return input.placeholder || input.name || '';
  }
  
  alert('Sending to API...');
  
  fetch(API_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ fields })
  })
  .then(r => r.json())
  .then(data => {
    forms.forEach(form => {
      const inputs = form.querySelectorAll('input, textarea, select');
      inputs.forEach(input => {
        const key = input.name || input.id;
        if(data[key]) {
          if(input.type === 'checkbox') {
            input.checked = data[key];
          } else if(input.type === 'radio') {
            if(input.value === data[key]) input.checked = true;
          } else {
            input.value = data[key];
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
          }
        }
      });
    });
    alert('Form filled!');
  })
  .catch(e => alert('Error: ' + e.message));
})();

