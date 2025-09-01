import { Routes } from '@angular/router';

<<<<<<< HEAD
import { HomePage } from '@/pages/home/home.page';
import { LoginPage } from '@/pages/login/login.page';
import { RegisterPage } from '@/pages/register/register.page';
import { ForgotPasswordPage } from '@/pages/forgot-password/forgot-password.page';
import { ResetPasswordPage } from '@/pages/reset-password/reset-password.page';
import { OAuthCallbackPage } from '@/pages/oauth-callback/oauth-callback.page';
import { BrowsePage } from '@/pages/browse/browse.page';
import { guestGuard } from '@/auth/guards/guest.guard';
import { authGuard } from '@/auth/guards/auth.guard';

export const routes: Routes = [
  { path: '', component: HomePage, canActivate: [guestGuard] },
  { path: 'login', component: LoginPage, canActivate: [guestGuard] },
  { path: 'register', component: RegisterPage, canActivate: [guestGuard] },
  { path: 'forgot-password', component: ForgotPasswordPage, canActivate: [guestGuard] },
  { path: 'reset-password', component: ResetPasswordPage, canActivate: [guestGuard] },
  { path: 'oauth/callback', component: OAuthCallbackPage, canActivate: [guestGuard] },

  // Protected
  { path: 'browse', component: BrowsePage, canActivate: [authGuard] },
];
=======
export const routes: Routes = [];
>>>>>>> 669195e367f919c154b0c3d6f114408079d8fd77
