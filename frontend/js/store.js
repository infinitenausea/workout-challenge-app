class Store {
  constructor() {
    this.state = {
      exercises: [],
      challenges: [],
      currentRoute: 'dashboard',
      currentChallengeId: null,
      currentChallenge: null,
      workouts: [],
      achievements: [],
      theme: 'dark'
    };
    this.listeners = [];
  }

  // Subscribe to changes. Returns unsubscribe function.
  subscribe(listener) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  // Update state and notify listeners
  setState(newState) {
    this.state = {
      ...this.state,
      ...newState
    };
    this.notify();
  }

  notify() {
    const currentListeners = [...this.listeners];
    currentListeners.forEach(listener => {
      if (this.listeners.includes(listener)) {
        listener(this.state);
      }
    });
  }

  getState() {
    return this.state;
  }

  // Action: Set exercises
  setExercises(exercises) {
    this.setState({ exercises });
  }

  // Action: Add a single exercise
  addExercise(exercise) {
    this.setState({
      exercises: [...this.state.exercises, exercise]
    });
  }

  // Action: Set challenges
  setChallenges(challenges) {
    this.setState({ challenges });
  }

  // Action: Add a single challenge
  addChallenge(challenge) {
    this.setState({
      challenges: [...this.state.challenges, challenge]
    });
  }

  // Action: Remove a challenge by id
  removeChallenge(challengeId) {
    this.setState({
      challenges: this.state.challenges.filter(c => c.id !== challengeId),
      currentChallengeId: this.state.currentChallengeId === challengeId ? null : this.state.currentChallengeId,
      currentChallenge: this.state.currentChallenge && this.state.currentChallenge.id === challengeId ? null : this.state.currentChallenge
    });
  }

  // Action: Set current challenge detail
  setCurrentChallenge(challenge) {
    this.setState({ currentChallenge: challenge });
  }

  // Action: Set workouts for the current challenge
  setWorkouts(workouts) {
    this.setState({ workouts });
  }

  // Action: Add workout to the beginning (DESC sorting)
  addWorkout(workout) {
    this.setState({
      workouts: [workout, ...this.state.workouts]
    });
  }

  // Action: Remove workout by id
  removeWorkout(workoutId) {
    this.setState({
      workouts: this.state.workouts.filter(w => w.id !== workoutId)
    });
  }

  // Action: Update progress of a specific challenge
  updateChallengeProgress(challengeId, newProgress, newStatus) {
    // Update in challenges list
    const updatedChallenges = this.state.challenges.map(c => {
      if (c.id === challengeId) {
        return {
          ...c,
          current_progress: newProgress,
          status: newStatus
        };
      }
      return c;
    });

    // Update currentChallenge if it matches
    let updatedCurrentChallenge = this.state.currentChallenge;
    if (updatedCurrentChallenge && updatedCurrentChallenge.id === challengeId) {
      updatedCurrentChallenge = {
        ...updatedCurrentChallenge,
        current_progress: newProgress,
        status: newStatus
      };
    }

    this.setState({
      challenges: updatedChallenges,
      currentChallenge: updatedCurrentChallenge
    });
  }

  // Action: Update details of a challenge in the list
  updateChallengeInList(updatedChallenge) {
    const updatedChallenges = this.state.challenges.map(c => 
      c.id === updatedChallenge.id ? updatedChallenge : c
    );

    let updatedCurrentChallenge = this.state.currentChallenge;
    if (updatedCurrentChallenge && updatedCurrentChallenge.id === updatedChallenge.id) {
      updatedCurrentChallenge = updatedChallenge;
    }

    this.setState({
      challenges: updatedChallenges,
      currentChallenge: updatedCurrentChallenge
    });
  }

  // Action: Set achievements (can be achievement objects or codes)
  setAchievements(achievements) {
    const codes = achievements.map(a => typeof a === 'object' ? a.achievement_code : a);
    this.setState({ achievements: codes });
  }

  // Action: Add newly unlocked achievements
  addAchievements(newCodes) {
    this.setState({
      achievements: [...this.state.achievements, ...newCodes]
    });
  }

  // Action: Navigate to route
  navigate(route, params = {}) {
    this.setState({
      currentRoute: route,
      ...params
    });
  }
}

export const store = new Store();

