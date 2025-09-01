import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';

type User = {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
};

type AuthResponse = {
  token: string; // JWT
  user: User;
};

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);

  private userSig = signal<User | null>(null);
  private tokenSig = signal<string | null>(null);

  readonly user = computed(() => this.userSig());
  readonly isAuthenticated = computed(() => !!this.tokenSig());

  constructor() {
    // Load from storage (browser-only)
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

  async login(credentials: { username: string; password: string }) {
    const res = await this.http.post<AuthResponse>('/api/auth/login', credentials).toPromise();
    this.setAuth(res!);
  }

  async register(payload: {
    email: string;
    username: string;
    firstName: string;
    lastName: string;
    password: string;
  }) {
    const res = await this.http.post<AuthResponse>('/api/auth/register', payload).toPromise();
    this.setAuth(res!);
  }

  async requestPasswordReset(email: string) {
    await this.http.post('/api/auth/forgot-password', { email }).toPromise();
  }

  async resetPassword(token: string, newPassword: string) {
    await this.http.post('/api/auth/reset-password', { token, password: newPassword }).toPromise();
  }

  async loadMe() {
    const me = await this.http.get<User>('/api/auth/me').toPromise();
    if (me) {
      this.userSig.set(me);
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_user', JSON.stringify(me));
      }
    }
  }

  logout() {
    this.userSig.set(null);
    this.tokenSig.set(null);
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
    }
    this.router.navigate(['/login']);
  }

  startOAuth(provider: '42' | 'google' | 'github') {
    if (typeof window !== 'undefined') {
      window.location.href = `/api/auth/${provider}`;
    }
  }

  completeOAuth(token: string) {
    // Backend should send back JWT via redirect to /oauth/callback?token=...
    this.tokenSig.set(token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', token);
    }
    // Fetch user
    this.loadMe().then(() => this.router.navigate(['/browse']));
  }

  private setAuth(res: AuthResponse) {
    this.tokenSig.set(res.token);
    this.userSig.set(res.user);
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', res.token);
      localStorage.setItem('auth_user', JSON.stringify(res.user));
    }
    this.router.navigate(['/browse']);
  }
}
