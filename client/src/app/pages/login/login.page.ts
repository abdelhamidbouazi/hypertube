import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';
import { MatIconModule } from '@angular/material/icon';

@Component({
  standalone: true,
  selector: 'app-login',
  imports: [CommonModule, FormsModule, RouterLink, MatIconModule],
  templateUrl: './login.page.html',
})
export class LoginPage {
  private auth = inject(AuthService);

  email = '';
  password = '';

  isLoading = false;
  loadingProvider: 'local' | 'google' | '42' | null = null;

  ngOnInit() {
    if (typeof window !== 'undefined' && window.history) {
      const state = history.state;

      if (state && state.fromRegistration) {
        this.email = state.email || '';
        this.password = state.password || '';
        console.log('Registration successful! Please login with your credentials.');
        console.log('Pre-filled email:', this.email); // Debug log
      }
    }
  }

  async submit() {
    if (this.isLoading) return;
    this.isLoading = true;
    this.loadingProvider = 'local';
    try {
      await this.auth.login({ email: this.email, password: this.password });
    } catch (err) {
      console.error('login failed', err);
    } finally {
      this.isLoading = false;
      this.loadingProvider = null;
    }
  }

  oauth(provider: '42' | 'google') {
    if (this.isLoading) return;
    this.isLoading = true;
    this.loadingProvider = provider;

    // window.location.href = `${environment.apiUrl}/auth/${provider}`;
  }
}
