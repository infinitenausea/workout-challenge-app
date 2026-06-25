import { store } from '../../store.js';
import { api } from '../../api.js';
import { showAchievementPopup } from './achievement-popup.js';

export class WorkoutModal {
  constructor() {
    this.overlay = null;
    this.challengeId = null;
    this.handleKeyDown = this.handleKeyDown.bind(this);
  }

  open(challengeId) {
    this.challengeId = challengeId;

    // Create overlay
    this.overlay = document.createElement('div');
    this.overlay.className = 'modal-overlay';

    const today = new Date().toISOString().split('T')[0];

    this.overlay.innerHTML = `
      <div class="modal-content">
        <div class="modal-header">
          <h3>Добавить тренировку</h3>
          <button id="modal-close-x" class="modal-close-btn">&times;</button>
        </div>
        <form id="workout-form" style="display: flex; flex-direction: column; gap: 16px; margin-top: 16px;">
          <div class="form-group" style="margin-bottom: 0;">
            <label for="workout_date">Дата тренировки</label>
            <input type="date" id="workout_date" name="workout_date" value="${today}" required>
          </div>
          <div class="form-group" style="margin-bottom: 0;">
            <label for="workout_value">Количество повторений</label>
            <input type="number" id="workout_value" name="value" min="1" placeholder="Например: 50" required>
          </div>
          <div style="display: flex; gap: 12px; margin-top: 8px;">
            <button type="submit" id="workout-submit-btn" style="flex: 1;">Сохранить</button>
            <button type="button" id="modal-cancel-btn" class="secondary" style="flex: 1;">Отмена</button>
          </div>
        </form>
      </div>
    `;

    document.body.appendChild(this.overlay);

    // Bind events
    this.overlay.querySelector('#workout-form').addEventListener('submit', (e) => this.handleSubmit(e));
    this.overlay.querySelector('#modal-close-x').addEventListener('click', () => this.close());
    this.overlay.querySelector('#modal-cancel-btn').addEventListener('click', () => this.close());
    
    // Close on overlay click
    this.overlay.addEventListener('click', (e) => {
      if (e.target === this.overlay) this.close();
    });

    // Close on Escape key
    document.addEventListener('keydown', this.handleKeyDown);
  }

  close() {
    if (this.overlay) {
      this.overlay.remove();
      this.overlay = null;
    }
    document.removeEventListener('keydown', this.handleKeyDown);
  }

  handleKeyDown(e) {
    if (e.key === 'Escape') {
      this.close();
    }
  }

  showToast(message, type = 'success') {
    const toastContainer = document.getElementById('toast-container');
    if (!toastContainer) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <span>${message}</span>
      <button style="background: none; border: none; color: inherit; font-size: 20px; cursor: pointer; padding: 0 4px;" onclick="this.parentElement.remove()">&times;</button>
    `;

    toastContainer.appendChild(toast);
    setTimeout(() => {
      toast.remove();
    }, 4000);
  }

  async handleSubmit(e) {
    e.preventDefault();

    const form = e.target;
    const value = parseInt(form.elements['value'].value, 10);
    const workout_date = form.elements['workout_date'].value;

    if (isNaN(value) || value <= 0) {
      this.showToast('Количество должно быть больше 0', 'error');
      return;
    }
    if (!workout_date) {
      this.showToast('Выберите дату тренировки', 'error');
      return;
    }

    const submitBtn = form.querySelector('#workout-submit-btn');
    submitBtn.disabled = true;

    try {
      const response = await api.createWorkout(this.challengeId, { workout_date, value });
      
      if (response.success) {
        // Add workout to store
        store.addWorkout(response.workout);

        // Update progress in store
        const state = store.getState();
        const current = state.currentChallenge;
        if (current) {
          const newProgress = current.current_progress + response.workout.value;
          const newStatus = newProgress >= current.target_value ? 'completed' : current.status;
          store.updateChallengeProgress(this.challengeId, newProgress, newStatus);
        }

        // Close modal
        this.close();

        // Toast success
        this.showToast(`Тренировка добавлена! +${value} повторений`, 'success');

        // Check and display achievements
        if (response.unlocked_achievements && response.unlocked_achievements.length > 0) {
          // Add to store
          store.addAchievements(response.unlocked_achievements);
          // Show popups
          showAchievementPopup(response.unlocked_achievements);
        }
      }
    } catch (error) {
      console.error('Failed to create workout:', error);
      this.showToast(error.message || 'Ошибка сохранения тренировки', 'error');
      submitBtn.disabled = false;
    }
  }
}
