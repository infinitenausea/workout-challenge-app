const BASE_URL = '/api';

class ApiClient {
  constructor() {
    this.userId = 'default_user_1';
  }

  setUserId(userId) {
    this.userId = userId;
  }

  async _request(endpoint, options = {}) {
    const url = `${BASE_URL}${endpoint}`;
    
    // Set headers
    const headers = {
      'Content-Type': 'application/json',
      'X-User-Id': this.userId,
      ...(options.headers || {})
    };

    const config = {
      ...options,
      headers
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        // If error response has JSON body, attempt to extract error message
        let errorMsg = `HTTP error! Status: ${response.status}`;
        try {
          const errorData = await response.json();
          if (errorData && errorData.message) {
            errorMsg = errorData.message;
          }
        } catch (e) {
          // Response body is not JSON or could not be read, check text instead
          try {
            const errorText = await response.text();
            if (errorText) errorMsg = errorText.trim();
          } catch (e2) {}
        }
        
        const error = new Error(errorMsg);
        error.status = response.status;
        throw error;
      }

      // Handle 204 No Content or empty response
      if (response.status === 204) {
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error(`API request to ${endpoint} failed:`, error);
      throw error;
    }
  }

  // Exercises
  getExercises() {
    return this._request('/exercises');
  }

  createExercise(name) {
    return this._request('/exercises', {
      method: 'POST',
      body: JSON.stringify({ name })
    });
  }

  // Challenges
  getChallenges() {
    return this._request('/challenges');
  }

  getChallengeDetail(challengeId) {
    return this._request(`/challenges/${challengeId}`);
  }

  createChallenge(payload) {
    return this._request('/challenges', {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  }

  deleteChallenge(challengeId) {
    return this._request(`/challenges/${challengeId}`, {
      method: 'DELETE'
    });
  }

  // Workouts
  createWorkout(challengeId, payload) {
    return this._request(`/challenges/${challengeId}/workouts`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  }

  deleteWorkout(workoutId) {
    return this._request(`/workouts/${workoutId}`, {
      method: 'DELETE'
    });
  }

  // Achievements
  getAchievements() {
    return this._request('/achievements');
  }
}

export const api = new ApiClient();

