// Encapsulation of window.Telegram.WebApp to avoid crashes when running outside Telegram.
class TelegramService {
  constructor() {
    this.webApp = window.Telegram ? window.Telegram.WebApp : null;
    this.backButtonCallback = null;
  }

  isReady() {
    return !!this.webApp;
  }

  ready() {
    if (this.webApp) {
      this.webApp.ready();
    }
  }

  expand() {
    if (this.webApp) {
      this.webApp.expand();
    }
  }

  getInitData() {
    return this.webApp ? this.webApp.initData : '';
  }

  getUser() {
    return this.webApp && this.webApp.initDataUnsafe ? this.webApp.initDataUnsafe.user : null;
  }

  getThemeParams() {
    return this.webApp ? this.webApp.themeParams : {};
  }

  getColorScheme() {
    return this.webApp ? this.webApp.colorScheme : null;
  }

  onThemeChanged(callback) {
    if (this.webApp) {
      this.webApp.onEvent('themeChanged', callback);
    }
  }

  triggerImpact(style) {
    if (this.webApp && this.webApp.HapticFeedback) {
      this.webApp.HapticFeedback.impactOccurred(style);
    }
  }

  triggerNotification(type) {
    if (this.webApp && this.webApp.HapticFeedback) {
      this.webApp.HapticFeedback.notificationOccurred(type);
    }
  }

  triggerSelection() {
    if (this.webApp && this.webApp.HapticFeedback) {
      this.webApp.HapticFeedback.selectionChanged();
    }
  }

  showBackButton(onClick) {
    if (this.webApp && this.webApp.BackButton) {
      if (this.backButtonCallback) {
        this.webApp.BackButton.offClick(this.backButtonCallback);
      }
      this.backButtonCallback = onClick;
      this.webApp.BackButton.onClick(onClick);
      this.webApp.BackButton.show();
    }
  }

  hideBackButton() {
    if (this.webApp && this.webApp.BackButton) {
      this.webApp.BackButton.hide();
      if (this.backButtonCallback) {
        this.webApp.BackButton.offClick(this.backButtonCallback);
        this.backButtonCallback = null;
      }
    }
  }

  enableClosingConfirmation() {
    if (this.webApp && typeof this.webApp.enableClosingConfirmation === 'function') {
      this.webApp.enableClosingConfirmation();
    }
  }

  disableClosingConfirmation() {
    if (this.webApp && typeof this.webApp.disableClosingConfirmation === 'function') {
      this.webApp.disableClosingConfirmation();
    }
  }
}

export const tg = new TelegramService();

