// Encapsulation of window.Telegram.WebApp to avoid crashes when running outside Telegram.
class TelegramService {
  constructor() {
    this.webApp = window.Telegram ? window.Telegram.WebApp : null;
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
}

export const tg = new TelegramService();

