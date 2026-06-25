import { store } from './store.js';
import { api } from './api.js';
import { Router } from './router.js';
import { Dashboard } from './components/dashboard/dashboard.js';
import { ChallengeForm } from './components/challenge/challenge-form.js';

// Define the routes map
const routes = {
  'dashboard': Dashboard,
  'challenge-form': ChallengeForm
};

// Initialize app when DOM is fully loaded
document.addEventListener('DOMContentLoaded', async () => {
  const appContainer = document.getElementById('app');
  
  // 1. Initialize router
  const router = new Router(routes, appContainer);
  
  // 2. Initial routing
  store.navigate('dashboard');

  // 3. Fetch initial exercises from API and load into store
  try {
    const exercises = await api.getExercises();
    store.setExercises(exercises);
  } catch (error) {
    console.error('Failed to load initial exercises:', error);
    // Even if it fails, we fall back to empty list so the app doesn't crash
    store.setExercises([]);
  }

  // 4. Fetch initial challenges from API and load into store
  try {
    const challenges = await api.getChallenges();
    store.setChallenges(challenges);
  } catch (error) {
    console.error('Failed to load initial challenges:', error);
    // Even if it fails, we fall back to empty list so the app doesn't crash
    store.setChallenges([]);
  }
});
