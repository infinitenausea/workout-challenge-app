import { store } from './store.js';
import { api } from './api.js';
import { Router } from './router.js';
import { Dashboard } from './components/dashboard/dashboard.js';
import { ChallengeForm } from './components/challenge/challenge-form.js';
import { ChallengeDetail } from './components/challenge/challenge-detail.js';
import { tg } from './telegram.js';

// Define the routes map
const routes = {
  'dashboard': Dashboard,
  'challenge-form': ChallengeForm,
  'challenge-detail': ChallengeDetail
};

// Initialize app when DOM is fully loaded
document.addEventListener('DOMContentLoaded', async () => {
  // 0. Initialize Telegram WebApp SDK
  tg.ready();
  tg.expand();

  // If running inside Telegram, update the user badge and set user ID in api client
  const tgUser = tg.getUser();
  if (tgUser) {
    const badge = document.getElementById('user-badge');
    if (badge) {
      badge.textContent = tgUser.username || `${tgUser.first_name} ${tgUser.last_name || ''}`.trim() || String(tgUser.id);
    }
    api.setUserId(String(tgUser.id));
  }

  // Theme application logic
  const applyTheme = () => {
    const tgScheme = tg.getColorScheme();
    let isLight = false;
    
    if (tgScheme) {
      isLight = tgScheme === 'light';
    } else {
      // Fallback for local development outside Telegram
      isLight = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches;
    }

    if (isLight) {
      document.body.classList.add('light-theme');
    } else {
      document.body.classList.remove('light-theme');
    }
  };

  // Apply initial theme and listen to theme updates
  applyTheme();
  tg.onThemeChanged(applyTheme);

  // Also listen to browser prefers-color-scheme changes when running outside Telegram
  if (!tg.isReady() && window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', applyTheme);
  }

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

  // 5. Fetch initial achievements from API and load into store
  try {
    const achievements = await api.getAchievements();
    store.setAchievements(achievements);
  } catch (error) {
    console.error('Failed to load initial achievements:', error);
    store.setAchievements([]);
  }
});

