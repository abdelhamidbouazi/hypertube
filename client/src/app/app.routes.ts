import { Routes } from '@angular/router';

import { HomePage } from '@/pages/home/home.page';
import { LoginPage } from '@/pages/login/login.page';
import { RegisterPage } from '@/pages/register/register.page';
import { ForgotPasswordPage } from '@/pages/forgot-password/forgot-password.page';
import { ResetPasswordPage } from '@/pages/reset-password/reset-password.page';
import { OAuthCallbackPage } from '@/pages/oauth-callback/oauth-callback.page';
import { BrowsePage } from '@/pages/browse/browse.page';
import { DiscoverPage } from '@/pages/discover/discover.page';
import { guestGuard } from '@/auth/guards/guest.guard';
import { authGuard } from '@/auth/guards/auth.guard';
import { initGuard } from '@/auth/guards/init/init.guard';
import { LoadingComponent } from '@/components/loading/loading.component';

export const routes: Routes = [
  { path: 'loading', component: LoadingComponent },

  // Guest routes
  { path: '', component: HomePage, canActivate: [initGuard, guestGuard] },
  { path: 'login', component: LoginPage, canActivate: [initGuard, guestGuard] },
  { path: 'register', component: RegisterPage, canActivate: [initGuard, guestGuard] },
  { path: 'forgot-password', component: ForgotPasswordPage, canActivate: [initGuard, guestGuard] },
  { path: 'reset-password', component: ResetPasswordPage, canActivate: [initGuard, guestGuard] },
  { path: 'oauth-callback', component: OAuthCallbackPage, canActivate: [initGuard, guestGuard] },

  // Protected routes
  { path: 'browse', component: BrowsePage, canActivate: [initGuard, authGuard] },
  { path: 'discover', component: DiscoverPage },
];
