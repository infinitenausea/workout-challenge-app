import { store } from '../../store.js';

export class Dashboard {
  constructor(container) {
    this.container = container;
    this.unsubscribe = null;
  }

  mount() {
    this.unsubscribe = store.subscribe(() => this.render());
    this.render();
  }

  unmount() {
    if (this.unsubscribe) this.unsubscribe();
  }

  render() {
    const state = store.getState();
    
    if (state.currentRoute !== 'dashboard') {
      this.container.innerHTML = '';
      return;
    }

    const hasChallenges = state.challenges && state.challenges.length > 0;

    let challengesHTML = '';
    if (!hasChallenges) {
      challengesHTML = `
        <div class="card" style="text-align: center; padding: 40px 20px;">
          <div style="font-size: 48px; margin-bottom: 16px;">🏃‍♂️</div>
          <h2 style="margin-bottom: 8px;">Активные челленджи</h2>
          <p style="color: var(--tg-theme-hint-color); margin-bottom: 24px;">У вас пока нет активных челленджей. Создайте первый!</p>
          <button id="create-challenge-btn">Создать первый челлендж</button>
        </div>
      `;
    } else {
      const cardsHTML = state.challenges.map(c => {
        const progressPercent = Math.min(100, Math.round((c.current_progress / c.target_value) * 100)) || 0;
        
        // Format dates nicely
        const formatDate = (dateStr) => {
          if (!dateStr) return '';
          const d = new Date(dateStr);
          return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
        };

        const startDateFormatted = formatDate(c.start_date);
        const endDateFormatted = formatDate(c.end_date);

        return `
          <div class="card challenge-card" data-id="${c.id}">
            <div class="challenge-header">
              <span class="challenge-title">${c.name}</span>
              <span class="challenge-status ${c.status === 'failed' ? 'badge-failed' : ''}">${c.status === 'active' ? 'Активен' : c.status === 'completed' ? 'Завершен' : 'Провален'}</span>
            </div>
            <div class="challenge-dates">
              📅 ${startDateFormatted} — ${endDateFormatted}
            </div>
            <div class="progress-wrapper">
              <div class="progress-info">
                <span>Прогресс: <strong>${c.current_progress}</strong> из <strong>${c.target_value}</strong></span>
                <span class="progress-percent">${progressPercent}%</span>
              </div>
              <div class="progress-container">
                <div class="progress-bar" style="width: ${progressPercent}%;"></div>
              </div>
            </div>
          </div>
        `;
      }).join('');

      challengesHTML = `
        <div class="dashboard-header-row">
          <h2>Активные челленджи</h2>
          <button id="create-challenge-btn" class="btn-small">+ Создать</button>
        </div>
        <div class="challenges-list" style="margin-bottom: 16px;">
          ${cardsHTML}
        </div>
      `;
    }

    this.container.innerHTML = `
      ${challengesHTML}
    `;

    const btn = this.container.querySelector('#create-challenge-btn');
    if (btn) {
      btn.addEventListener('click', () => {
        store.navigate('challenge-form');
      });
    }

    // Attach click listeners to challenge cards
    this.container.querySelectorAll('.challenge-card').forEach(card => {
      card.addEventListener('click', (e) => {
        const id = parseInt(e.currentTarget.getAttribute('data-id'), 10);
        store.navigate('challenge-detail', { currentChallengeId: id });
      });
    });
  }
}

