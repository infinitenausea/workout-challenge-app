import { store } from './store.js';

export class Router {
  constructor(routes, container) {
    this.routes = routes;
    this.container = container;
    this.currentView = null;
    
    // Listen to store navigation events
    store.subscribe((state) => this.handleRoute(state.currentRoute));
  }

  handleRoute(route) {
    // If a view is already mounted, unmount it
    if (this.currentView && typeof this.currentView.unmount === 'function') {
      this.currentView.unmount();
    }

    const ViewClass = this.routes[route];
    if (!ViewClass) {
      console.error(`Route "${route}" not found. Falling back to dashboard.`);
      store.navigate('dashboard');
      return;
    }

    // Initialize and mount the new view
    this.container.innerHTML = '';
    this.currentView = new ViewClass(this.container);
    if (typeof this.currentView.mount === 'function') {
      this.currentView.mount();
    } else if (typeof this.currentView.render === 'function') {
      this.currentView.render();
    }
  }
}
