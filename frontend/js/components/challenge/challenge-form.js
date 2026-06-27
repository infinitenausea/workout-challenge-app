import { store } from '../../store.js';
import { api } from '../../api.js';
import { tg } from '../../telegram.js';

export class ChallengeForm {
  constructor(container) {
    this.container = container;
    this.unsubscribe = null;
    this.isSubmitting = false;
  }

  mount() {
    this.unsubscribe = store.subscribe(() => this.render());
    this.render();
  }

  unmount() {
    if (this.unsubscribe) {
      this.unsubscribe();
    }
    tg.disableClosingConfirmation();
    store.setState({ editMode: false, challengeId: null });
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

  async handleFormSubmit(e) {
    e.preventDefault();
    if (this.isSubmitting) return;

    const state = store.getState();
    const isEditMode = state.editMode || false;
    const challengeId = state.challengeId;

    const form = e.target;
    const name = form.elements['name'].value.trim();
    const exerciseIdVal = form.elements['exercise_id']?.value;
    const customExerciseName = form.elements['custom_exercise_name'] ? form.elements['custom_exercise_name'].value.trim() : '';
    const targetValue = parseInt(form.elements['target_value'].value, 10);
    const startDate = form.elements['start_date'].value;
    const endDate = form.elements['end_date'].value;

    // Client-side validations
    if (!name) {
      this.showToast('Введите название челленджа', 'error');
      return;
    }
    if (!isEditMode && exerciseIdVal === 'custom' && !customExerciseName) {
      this.showToast('Введите название нового упражнения', 'error');
      return;
    }
    if (isNaN(targetValue) || targetValue <= 0) {
      this.showToast('Целевое количество должно быть больше 0', 'error');
      return;
    }
    if (new Date(endDate) < new Date(startDate)) {
      this.showToast('Дата окончания не может быть раньше даты старта', 'error');
      return;
    }

    this.isSubmitting = true;
    const submitBtn = form.querySelector('button[type="submit"]');
    if (submitBtn) submitBtn.disabled = true;

    try {
      if (isEditMode) {
        const challengePayload = {
          name,
          target_value: targetValue,
          start_date: `${startDate}T00:00:00Z`,
          end_date: `${endDate}T00:00:00Z`
        };

        console.log('Challenge payload to update:', challengePayload);
        
        const updatedChallenge = await api.updateChallenge(challengeId, challengePayload);
        store.updateChallengeInList(updatedChallenge);
        this.showToast('Челлендж успешно изменен!', 'success');
        tg.disableClosingConfirmation();

        // Navigate back to details
        store.navigate('challenge-detail', { currentChallengeId: challengeId });
      } else {
        let finalExerciseId = parseInt(exerciseIdVal, 10);

        // Handle custom exercise creation first
        if (exerciseIdVal === 'custom') {
          try {
            const newExercise = await api.createExercise(customExerciseName);
            this.showToast(`Упражнение "${newExercise.name}" успешно создано!`, 'success');
            
            // Add to store
            store.addExercise(newExercise);
            finalExerciseId = newExercise.id;
          } catch (error) {
            this.showToast(error.message || 'Ошибка создания упражнения', 'error');
            this.isSubmitting = false;
            if (submitBtn) submitBtn.disabled = false;
            return;
          }
        }

        // Prepare challenge payload with RFC3339 dates
        const challengePayload = {
          name,
          exercise_id: finalExerciseId,
          target_value: targetValue,
          start_date: `${startDate}T00:00:00Z`,
          end_date: `${endDate}T00:00:00Z`
        };

        console.log('Challenge payload to save:', challengePayload);
        
        const newChallenge = await api.createChallenge(challengePayload);
        store.addChallenge(newChallenge);
        this.showToast('Челлендж успешно создан!', 'success');
        tg.disableClosingConfirmation();

        // Navigate back to dashboard
        store.navigate('dashboard');
      }

    } catch (error) {
      this.showToast(error.message || 'Ошибка при сохранении челленджа', 'error');
    } finally {
      this.isSubmitting = false;
      if (submitBtn) submitBtn.disabled = false;
    }
  }

  render() {
    const state = store.getState();
    
    // We only render if we are currently on the 'challenge-form' route
    if (state.currentRoute !== 'challenge-form') {
      this.container.innerHTML = '';
      return;
    }

    const isEditMode = state.editMode || false;
    const challengeId = state.challengeId;
    let challengeToEdit = null;
    if (isEditMode && challengeId) {
      challengeToEdit = state.challenges.find(c => c.id === challengeId) || state.currentChallenge;
    }

    // Determine if custom exercise option is selected to persist toggle state
    // (though since it rerenders we can check the DOM if it exists, or just read the select value)
    let currentSelectVal = this.container.querySelector('#exercise_id')?.value || '';
    if (isEditMode && challengeToEdit && !currentSelectVal) {
      currentSelectVal = challengeToEdit.exercise_id;
    }

    // Check if we need to show the custom input field
    const showCustomInput = currentSelectVal === 'custom';

    const exercisesOptions = state.exercises.map(ex => 
      `<option value="${ex.id}" ${currentSelectVal == ex.id ? 'selected' : ''}>${ex.name}</option>`
    ).join('');

    // Default dates (today and today + 30 days)
    let today = new Date().toISOString().split('T')[0];
    let defaultEnd = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    let isStartDateDisabled = false;

    if (isEditMode && challengeToEdit) {
      today = challengeToEdit.start_date.split('T')[0];
      defaultEnd = challengeToEdit.end_date.split('T')[0];

      // Check if original start_date is today or in the past
      const now = new Date();
      const todayDate = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const originalStart = new Date(challengeToEdit.start_date);
      const originalStartDay = new Date(originalStart.getFullYear(), originalStart.getMonth(), originalStart.getDate());
      if (originalStartDay <= todayDate) {
        isStartDateDisabled = true;
      }
    }

    this.container.innerHTML = `
      <div class="card challenge-form-card">
        <h2>${isEditMode ? 'Редактировать челлендж' : 'Новый челлендж'}</h2>
        <form id="challenge-form">
          <div class="form-group">
            <label for="name">Название челленджа</label>
            <input type="text" id="name" name="name" placeholder="Например: 100 отжиманий ежедневно" value="${isEditMode && challengeToEdit ? challengeToEdit.name : ''}" required>
          </div>

          <div class="form-group">
            <label for="exercise_id">Упражнение</label>
            <select id="exercise_id" name="exercise_id" required ${isEditMode ? 'disabled' : ''}>
              <option value="" disabled ${!currentSelectVal ? 'selected' : ''}>Выберите упражнение...</option>
              ${exercisesOptions}
              ${isEditMode ? '' : '<option value="custom" ' + (currentSelectVal === 'custom' ? 'selected' : '') + '>+ Добавить свое...</option>'}
            </select>
          </div>

          <div class="form-group" id="custom-exercise-group" style="display: ${showCustomInput ? 'block' : 'none'};">
            <label for="custom_exercise_name">Название нового упражнения</label>
            <input type="text" id="custom_exercise_name" name="custom_exercise_name" placeholder="Например: Планка">
          </div>

          <div class="form-group">
            <label for="target_value">Цель (повторений)</label>
            <input type="number" id="target_value" name="target_value" min="1" placeholder="3000" value="${isEditMode && challengeToEdit ? challengeToEdit.target_value : ''}" required>
          </div>

          <div class="form-group-row" style="display: flex; gap: 12px; margin-bottom: 20px;">
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
              <label for="start_date">Старт</label>
              <input type="date" id="start_date" name="start_date" value="${today}" required ${isStartDateDisabled ? 'disabled' : ''}>
            </div>
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
              <label for="end_date">Дедлайн</label>
              <input type="date" id="end_date" name="end_date" value="${defaultEnd}" required>
            </div>
          </div>

          <div style="display: flex; gap: 12px;">
            <button type="submit" style="flex: 1;">${isEditMode ? 'Сохранить' : 'Создать'}</button>
            <button type="button" id="cancel-btn" class="secondary" style="flex: 1;">Отмена</button>
          </div>
        </form>
      </div>
    `;

    // Attach listeners
    const form = this.container.querySelector('#challenge-form');
    form.addEventListener('submit', (e) => this.handleFormSubmit(e));
    form.addEventListener('input', () => {
      tg.enableClosingConfirmation();
    });

    const select = this.container.querySelector('#exercise_id');
    const customGroup = this.container.querySelector('#custom-exercise-group');
    
    if (select) {
      select.addEventListener('change', (e) => {
        if (e.target.value === 'custom') {
          customGroup.style.display = 'block';
          customGroup.querySelector('input').setAttribute('required', 'required');
        } else {
          customGroup.style.display = 'none';
          customGroup.querySelector('input').removeAttribute('required');
        }
      });
    }

    const cancelBtn = this.container.querySelector('#cancel-btn');
    cancelBtn.addEventListener('click', () => {
      tg.disableClosingConfirmation();
      if (isEditMode) {
        store.navigate('challenge-detail', { currentChallengeId: challengeId });
      } else {
        store.navigate('dashboard');
      }
    });
  }
}
