// Initialize Particles.js
particlesJS('particles-js', {
  particles: {
    number: { value: 80, density: { enable: true, value_area: 800 } },
    color: { value: "#ffffff" },
    shape: { type: "circle" },
    opacity: { value: 0.5 },
    size: { value: 3, random: true },
    move: { enable: true, speed: 2 },
    line_linked: { enable: true, distance: 150, color: "#ffffff" },
  },
});

// Fetch available COM ports on page load
window.addEventListener('DOMContentLoaded', async () => {
  const response = await fetch('/api/ports');
  const ports = await response.json();
  const comPortSelect = document.getElementById('comPort');

  ports.forEach(port => {
    const option = document.createElement('option');
    option.value = port;
    option.textContent = port;
    comPortSelect.appendChild(option);
  });
});

// Handle form submission and start scan
document.getElementById('configForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const config = {
    comPort: document.getElementById('comPort').value,
    baudRate: parseInt(document.getElementById('baudRate').value),
    parity: document.getElementById('parity').value,
    slaveId: parseInt(document.getElementById('slaveId').value),
    startRegister: parseInt(document.getElementById('startRegister').value),
    numRegisters: parseInt(document.getElementById('numRegisters').value),
  };

  const response = await fetch('/api/scan', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  });

  const data = await response.json();
  const resultsDiv = document.getElementById('results');
  resultsDiv.innerHTML = '<h3 class="text-light">Scan Results:</h3>';

  data.forEach(result => {
    const p = document.createElement('p');
    p.textContent = result;
    p.classList.add('text-light');
    resultsDiv.appendChild(p);
  });
});
