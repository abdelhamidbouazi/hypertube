// filepath: /src/app/pages/login/login.page.ts
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
  email = 'example@example.org';
  password = '!!!!!!!!!!!';

  submit() {
    this.auth.login({ email: this.email, password: this.password });
  }
  oauth(provider: '42' | 'google' | 'github') {
    this.auth.startOAuth(provider);
  }
}
