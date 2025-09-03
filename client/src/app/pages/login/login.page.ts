import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';
import { MatIconModule } from '@angular/material/icon';
import { environment } from '~/environments/environment';

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
