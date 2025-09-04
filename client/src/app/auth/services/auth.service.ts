import { Injectable, inject, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import axios from 'axios';
import { environment } from '~/environments/environment';

type User = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
};

type AuthResponse = {
  token: string;
  user: User;
};

type BackendAuthResponse = {
  AccessToken: string;
  ExpiresIn: number;
  RefreshToken: string;
  TokenType: string;
};

@Injectable({ providedIn: 'root' })
export class AuthService {
  private router = inject(Router);

  private userSig = signal<User | null>(null);
  private tokenSig = signal<string | null>(null);

  readonly user = computed(() => this.userSig());
  readonly isAuthenticated = computed(() => !!this.tokenSig());

  constructor() {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('auth_token');
      const userJson = localStorage.getItem('auth_user');
      if (token && userJson) {
        this.tokenSig.set(token);
        this.userSig.set(JSON.parse(userJson));
      }
    }
  }

  get token(): string | null {
    return this.tokenSig();
  }

  async login(credentials: { email: string; password: string }) {
    try {
      const response = await axios.post<BackendAuthResponse>(
        `${environment.apiUrl}/auth/login`,
        credentials,
        { headers: { 'Content-Type': 'application/json' } }
      );

      console.log('Login response:', response.data);
      await this.setAuthFromResponse(response.data);
    } catch (err: unknown) {
      console.error('login failed', err);
      throw err;
    }
  }

  async register(payload: {
    email: string;
    firstname: string;
    lastname: string;
    password: string;
  }) {
    try {
      console.log('Register payload:', payload);
      const response = await axios.post<BackendAuthResponse>(
        `${environment.apiUrl}/auth/register`,
        payload,
        { headers: { 'Content-Type': 'application/json' } }
      );
      console.log('Registration response:', response.data);
      // await this.setAuthFromResponse(response.data);
    } catch (err: unknown) {
      console.error('registration failed', err);
      throw err;
    }
  }

  async requestPasswordReset(email: string) {
    try {
      await axios.post(
        `${environment.apiUrl}/auth/forgot-password`,
        { email },
        { headers: { 'Content-Type': 'application/json' } }
      );
    } catch (err: unknown) {
      console.error('request password reset failed', err);
      throw err;
    }
  }

  async resetPassword(token: string, newPassword: string) {
    try {
      await axios.post(
        `${environment.apiUrl}/auth/reset-password`,
        { token, password: newPassword },
        { headers: { 'Content-Type': 'application/json' } }
      );
    } catch (err: unknown) {
      console.error('reset password failed', err);
      throw err;
    }
  }

  async loadMe() {
    const token =
      this.tokenSig() ??
      (typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null);
    const headers: Record<string, string> = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const response = await axios.get<User>(`${environment.apiUrl}/auth/me`, { headers });
    const me = response.data;
    if (me) {
      this.userSig.set(me);
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_user', JSON.stringify(me));
      }
    }
  }

  // logout() {
  //   this.userSig.set(null);
  //   this.tokenSig.set(null);
  //   if (typeof window !== 'undefined') {
  //     localStorage.removeItem('auth_token');
  //     localStorage.removeItem('auth_user');
  //   }
  //   this.router.navigate(['/login']);
  // }

  // startOAuth(provider: '42' | 'google' | 'github') {
  //   if (typeof window !== 'undefined') {
  //     window.location.href = `${environment.apiUrl}/auth/${provider}`;
  //   }
  // }

  // completeOAuth(token: string) {
  //   this.tokenSig.set(token);
  //   if (typeof window !== 'undefined') {
  //     localStorage.setItem('auth_token', token);
  //   }

  //   this.loadMe().then(() => this.router.navigate(['/browse']));
  // }

  // private setAuth(res: AuthResponse) {
  //   this.tokenSig.set(res.token);
  //   this.userSig.set(res.user);
  //   if (typeof window !== 'undefined') {
  //     localStorage.setItem('auth_token', res.token);
  //     localStorage.setItem('auth_user', JSON.stringify(res.user));
  //   }
  //   this.router.navigate(['/browse']);
  // }

  private async setAuthFromResponse(res: BackendAuthResponse) {
    this.tokenSig.set(res.AccessToken);
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', res.AccessToken);
      localStorage.setItem('refresh_token', res.RefreshToken);
    }
    try {
      // await this.loadMe();
      console.log('setAuthFromResponse: calling loadMe');
    } catch (e) {
      console.warn('loadMe failed', e);
    }
    this.router.navigate(['/browse']);
  }
}
