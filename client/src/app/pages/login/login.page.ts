// filepath: /src/app/pages/login/login.page.ts
import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-login',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './login.page.html',
})
export class LoginPage {
  private auth = inject(AuthService);
  username = '';
  password = '';

  submit() {
    this.auth.login({ username: this.username, password: this.password });
  }
  oauth(provider: '42' | 'google' | 'github') {
    this.auth.startOAuth(provider);
  }
}
