const achievementData = {
  first_step: { icon: '🌱', name: 'Первый шаг', desc: 'Внесена первая тренировка' },
  equator: { icon: '📈', name: 'Экватор', desc: 'Прогресс достиг 50%' },
  hero: { icon: '⚡', name: 'Герой', desc: 'Челлендж завершён до дедлайна!' },
  stability: { icon: '🔥', name: 'Стабильность', desc: 'Тренировки 3 дня подряд' }
};

let popupQueue = [];
let isShowingPopup = false;

export function showAchievementPopup(achievementCodes) {
  if (!achievementCodes || achievementCodes.length === 0) return;
  
  popupQueue.push(...achievementCodes);
  processQueue();
}

function processQueue() {
  if (isShowingPopup || popupQueue.length === 0) return;

  isShowingPopup = true;
  const code = popupQueue.shift();

  displayPopup(code, () => {
    isShowingPopup = false;
    processQueue();
  });
}

function displayPopup(code, callback) {
  const data = achievementData[code];
  if (!data) {
    callback();
    return;
  }

  const overlay = document.createElement('div');
  overlay.className = 'achievement-popup-overlay';

  overlay.innerHTML = `
    <div class="achievement-popup-card">
      <div class="achievement-icon">${data.icon}</div>
      <h2 style="margin-bottom: 8px; color: #FFD700;">${data.name}</h2>
      <p style="margin-bottom: 24px; color: var(--tg-theme-text-color); font-size: 15px;">${data.desc}</p>
      <button id="achievement-ok-btn" style="width: 100%; background: linear-gradient(135deg, #FFD700, #FFA500);">Отлично!</button>
      <div class="achievement-auto-close-bar"></div>
    </div>
  `;

  document.body.appendChild(overlay);

  let cleanedUp = false;
  
  const closePopup = () => {
    if (cleanedUp) return;
    cleanedUp = true;
    
    clearTimeout(autoCloseTimeout);
    overlay.remove();
    callback();
  };

  // Auto close after 5 seconds
  const autoCloseTimeout = setTimeout(closePopup, 5000);

  // Bind clicks
  overlay.querySelector('#achievement-ok-btn').addEventListener('click', closePopup);
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) closePopup();
  });
}
