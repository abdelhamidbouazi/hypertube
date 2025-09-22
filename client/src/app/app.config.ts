import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
  provideZonelessChangeDetection,
} from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from '@/app.routes';
import { provideClientHydration, withEventReplay } from '@angular/platform-browser';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { provideNzI18n, en_US } from 'ng-zorro-antd/i18n';
import { provideNzIcons } from 'ng-zorro-antd/icon';
import {
  SearchOutline,
  BellOutline,
  UserOutline,
  PlusOutline,
  LikeOutline,
  SoundOutline,
  LeftOutline,
  RightOutline,
  StarFill,
} from '@ant-design/icons-angular/icons';
import { registerLocaleData } from '@angular/common';
import en from '@angular/common/locales/en';

registerLocaleData(en);

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideRouter(routes),
    provideClientHydration(withEventReplay()),
    provideAnimations(),
    provideHttpClient(withFetch()),
    provideNzI18n(en_US),
    provideNzIcons([
      SearchOutline,
      BellOutline,
      UserOutline,
      PlusOutline,
      LikeOutline,
      SoundOutline,
      LeftOutline,
      RightOutline,
      StarFill,
    ]),
  ],
};
