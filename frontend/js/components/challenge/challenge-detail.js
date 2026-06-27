import { store } from '../../store.js';
import { api } from '../../api.js';
import { WorkoutModal } from '../ui/workout-modal.js';
import { tg } from '../../telegram.js';
import { showAchievementInfo } from '../ui/achievement-popup.js';

export class ChallengeDetail {
  constructor(container) {
    this.container = container;
    this.unsubscribe = null;
  }

  async mount() {
    this.unsubscribe = store.subscribe(() => this.render());

    // Clear previous challenge detail to avoid flash of old content
    store.setCurrentChallenge(null);
    store.setWorkouts([]);
    store.setCurrentChallengeAchievements([]);

    const state = store.getState();
    const challengeId = state.currentChallengeId;

    if (challengeId) {
      try {
        const detail = await api.getChallengeDetail(challengeId);
        store.setCurrentChallenge(detail);
        store.setWorkouts(detail.workouts || []);
        
        try {
          const achievements = await api.getChallengeAchievements(challengeId);
          store.setCurrentChallengeAchievements(achievements);
        } catch (achErr) {
          console.error('Failed to load challenge achievements:', achErr);
          store.setCurrentChallengeAchievements([]);
        }
      } catch (error) {
        console.error('Failed to load challenge details:', error);
        this.showToast('Ошибка при загрузке деталей челленджа', 'error');
      }
    }
  }

  unmount() {
    if (this.unsubscribe) {
      this.unsubscribe();
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

  async handleDeleteWorkout(workoutId) {
    if (!confirm('Вы уверены, что хотите удалить эту тренировку?')) return;

    tg.triggerImpact('medium');
    try {
      const state = store.getState();
      const challengeId = state.currentChallengeId;
      
      const response = await api.deleteWorkout(workoutId);
      if (response.success) {
        store.removeWorkout(workoutId);
        store.updateChallengeProgress(
          challengeId,
          response.challenge.current_progress,
          response.challenge.status
        );
        this.showToast('Тренировка успешно удалена', 'success');
      }
    } catch (error) {
      console.error('Failed to delete workout:', error);
      this.showToast(error.message || 'Ошибка удаления тренировки', 'error');
    }
  }

  async handleDeleteChallenge(challengeId) {
    if (!confirm('Удалить челлендж? Это действие необратимо — все тренировки будут удалены.')) return;

    try {
      await api.deleteChallenge(challengeId);
      store.removeChallenge(challengeId);
      tg.triggerNotification('success');
      store.navigate('dashboard');
      this.showToast('Челлендж удалён', 'success');
    } catch (error) {
      console.error('Failed to delete challenge:', error);
      this.showToast(error.message || 'Ошибка удаления челленджа', 'error');
    }
  }

  render() {
    const state = store.getState();

    if (state.currentRoute !== 'challenge-detail' || !state.currentChallenge) {
      this.container.innerHTML = '';
      return;
    }

    const c = state.currentChallenge;
    const progressPercent = Math.min(100, Math.round((c.current_progress / c.target_value) * 100)) || 0;

    // Format dates
    const formatDate = (dateStr) => {
      if (!dateStr) return '';
      const d = new Date(dateStr);
      return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
    };

    // Calculate countdown
    let timerText = '';
    if (c.status === 'completed') {
      timerText = 'Челлендж завершён! 🎉';
    } else {
      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const end = new Date(c.end_date);
      const endDay = new Date(end.getFullYear(), end.getMonth(), end.getDate());

      const diffTime = endDay - today;
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

      if (diffDays > 0) {
        timerText = `Осталось: ${diffDays} дн.`;
      } else if (diffDays === 0) {
        timerText = 'Последний день!';
      } else {
        timerText = 'Дедлайн истёк';
      }
    }

    // Achievements grid mapping
    const allAchievements = [
      { code: 'first_step', icon: '🌱', name: 'Первый шаг' },
      { code: 'equator', icon: '📈', name: 'Экватор' },
      { code: 'hero', icon: '⚡', name: 'Герой' },
      { code: 'stability', icon: '🔥', name: 'Стабильность' },
      { code: 'power_start', icon: '🚀', name: 'Ударный старт' },
      { code: 'overachiever', icon: '🏆', name: 'Перевыполнение' },
      { code: 'early_bird', icon: '🌅', name: 'Ранняя пташка' },
      { code: 'final_spurt', icon: '🏁', name: 'Финальный рывок' }
    ];

    const currentAchCodes = state.currentChallengeAchievements || [];
    const achievementsHTML = `
      <div class="achievements-grid">
        ${allAchievements.map(ach => {
          const isUnlocked = currentAchCodes.includes(ach.code);
          return `
            <div class="achievement-grid-item ${isUnlocked ? '' : 'locked'}" data-code="${ach.code}">
              <div class="ach-icon">${ach.icon}</div>
              <div class="ach-name">${ach.name}</div>
            </div>
          `;
        }).join('')}
      </div>
    `;

    // Workouts list
    let workoutsListHTML = '';
    if (!state.workouts || state.workouts.length === 0) {
      workoutsListHTML = `<div style="text-align: center; color: var(--tg-theme-hint-color); padding: 20px;">Тренировок пока нет...</div>`;
    } else {
      workoutsListHTML = state.workouts.map(w => {
        const dateFormatted = formatDate(w.workout_date);
        return `
          <div class="workout-list-item" data-id="${w.id}">
            <div>
              <div style="font-weight: 600;">+${w.value} повторений</div>
              <div style="font-size: 12px; color: var(--tg-theme-hint-color);">${dateFormatted}</div>
            </div>
            <button class="delete-workout-btn" data-id="${w.id}" style="background: none; border: none; color: var(--danger-color); cursor: pointer; padding: 8px;">
              🗑️
            </button>
          </div>
        `;
      }).join('');
    }

    const isInactive = c.status === 'completed' || c.status === 'failed';

    this.container.innerHTML = `
      <div class="challenge-detail">
        <div class="challenge-detail-header">
          <button id="back-btn" class="secondary btn-small">← Назад</button>
          ${c.status === 'active' ? '<button id="edit-challenge-btn" class="secondary btn-small">✏️ Редактировать</button>' : ''}
          <span class="challenge-status ${c.status === 'failed' ? 'badge-failed' : ''}">${c.status === 'active' ? 'Активен' : c.status === 'completed' ? 'Завершен' : 'Провален'}</span>
        </div>

        <div class="card">
          <h2 style="margin-bottom: 8px;">${c.name}</h2>
          <div class="challenge-dates" style="margin-bottom: 16px;">
            📅 ${formatDate(c.start_date)} — ${formatDate(c.end_date)}
          </div>

          <div class="progress-wrapper">
            <div class="progress-info">
              <span>Прогресс: <strong>${c.current_progress}</strong> из <strong>${c.target_value}</strong></span>
              <span class="progress-percent">${progressPercent}%</span>
            </div>
            <div class="progress-container" style="height: 12px; background-color: var(--tg-theme-secondary-bg-color);">
              <div class="progress-bar" style="width: ${progressPercent}%;"></div>
            </div>
          </div>

          <div class="countdown-timer" style="margin-top: 16px; font-weight: 600; color: ${c.status === 'completed' ? 'var(--accent-color)' : 'var(--tg-theme-hint-color)'};">
            ${timerText}
          </div>

          <div style="margin-top: 16px; border-top: 1px solid var(--tg-theme-secondary-bg-color); padding-top: 16px;">
            <h4 style="margin-bottom: 12px; font-size: 14px; color: var(--tg-theme-hint-color);">Достижения челленджа</h4>
            ${achievementsHTML}
          </div>
        </div>

        <button id="add-workout-btn" class="add-workout-btn" ${isInactive ? 'disabled style="opacity: 0.5; cursor: not-allowed;"' : ''} style="width: 100%; margin-bottom: 24px;">
          Добавить тренировку
        </button>

        <h3>История тренировок</h3>
        <div class="workouts-list">
          ${workoutsListHTML}
        </div>

        <button id="delete-challenge-btn" class="danger" style="width: 100%; margin-top: 24px;">
          🗑️ Удалить челлендж
        </button>
      </div>
    `;

    // Bind events
    this.container.querySelector('#back-btn').addEventListener('click', () => {
      store.navigate('dashboard');
    });

    if (c.status === 'active') {
      const editBtn = this.container.querySelector('#edit-challenge-btn');
      if (editBtn) {
        editBtn.addEventListener('click', () => {
          store.navigate('challenge-form', { editMode: true, challengeId: c.id });
        });
      }
    }

    if (!isInactive) {
      this.container.querySelector('#add-workout-btn').addEventListener('click', () => {
        const modal = new WorkoutModal();
        modal.open(c.id);
      });
    }

    this.container.querySelectorAll('.delete-workout-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const id = parseInt(e.currentTarget.getAttribute('data-id'), 10);
        this.handleDeleteWorkout(id);
      });
    });

    this.container.querySelector('#delete-challenge-btn').addEventListener('click', () => {
      this.handleDeleteChallenge(c.id);
    });

    // Bind clicks to achievements
    this.container.querySelectorAll('.achievement-grid-item').forEach(item => {
      item.addEventListener('click', (e) => {
        const code = e.currentTarget.getAttribute('data-code');
        showAchievementInfo(code);
      });
    });
  }
}
