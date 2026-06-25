class Store {
  constructor() {
    this.state = {
      exercises: [],
      challenges: [],
      currentRoute: 'dashboard',
      currentChallengeId: null,
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

  // Action: Navigate to route
  navigate(route, params = {}) {
    this.setState({
      currentRoute: route,
      ...params
    });
  }
}

export const store = new Store();
