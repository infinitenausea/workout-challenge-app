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

    this.container.innerHTML = `
      <div class="card" style="text-align: center; padding: 40px 20px;">
        <div style="font-size: 48px; margin-bottom: 16px;">🏃‍♂️</div>
        <h2 style="margin-bottom: 8px;">Активные челленджи</h2>
        <p style="color: var(--tg-theme-hint-color); margin-bottom: 24px;">У вас пока нет активных челленджей. Создайте первый!</p>
        <button id="create-challenge-btn">Создать первый челлендж</button>
      </div>

      <div class="card">
        <h3 style="color: var(--tg-theme-link-color); margin-bottom: 16px; border-bottom: 1px solid var(--tg-theme-secondary-bg-color); padding-bottom: 8px;">Мои Достижения</h3>
        <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; text-align: center;">
          <div style="opacity: 0.3; filter: grayscale(100%);">
            <div style="font-size: 24px;">🌱</div>
            <div style="font-size: 10px; color: var(--tg-theme-hint-color);">Первый шаг</div>
          </div>
          <div style="opacity: 0.3; filter: grayscale(100%);">
            <div style="font-size: 24px;">📈</div>
            <div style="font-size: 10px; color: var(--tg-theme-hint-color);">Экватор</div>
          </div>
          <div style="opacity: 0.3; filter: grayscale(100%);">
            <div style="font-size: 24px;">⚡</div>
            <div style="font-size: 10px; color: var(--tg-theme-hint-color);">Герой</div>
          </div>
          <div style="opacity: 0.3; filter: grayscale(100%);">
            <div style="font-size: 24px;">🔥</div>
            <div style="font-size: 10px; color: var(--tg-theme-hint-color);">Стабильность</div>
          </div>
        </div>
      </div>
    `;

    const btn = this.container.querySelector('#create-challenge-btn');
    btn.addEventListener('click', () => {
      store.navigate('challenge-form');
    });
  }
}
