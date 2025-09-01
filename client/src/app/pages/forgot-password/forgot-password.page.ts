import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-forgot-password',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './forgot-password.page.html',
})
export class ForgotPasswordPage {
  private auth = inject(AuthService);
  email = '';

  submit() {
    this.auth.requestPasswordReset(this.email);
  }
}
