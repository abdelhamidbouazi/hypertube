import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';
import { MatIconModule } from '@angular/material/icon';

@Component({
  standalone: true,
  selector: 'app-forgot-password',
  imports: [CommonModule, FormsModule, RouterLink, MatIconModule],
  templateUrl: './forgot-password.page.html',
})
export class ForgotPasswordPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  email = '';
  emailError = '';
  formError = '';
  isSubmitting = false;

  validateEmail(): boolean {
    this.email = this.email.trim();
    if (!this.email) {
      this.emailError = 'Email is required';
      return false;
    }

    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(this.email)) {
      this.emailError = 'Please enter a valid email address';
      return false;
    }

    this.emailError = '';
    return true;
  }

  onEmailInput() {
    this.emailError = '';
  }

  private maskEmail(email: string): string {
    const [local, domain] = email.split('@');
    if (!domain) return email; // fallback, invalid but fine
    if (local.length <= 1) return `*@${domain}`;
    return `${local[0]}${'*'.repeat(Math.max(1, Math.min(6, local.length - 1)))}` + `@${domain}`;
  }

  private goToConfirmation(maskedEmail: string) {
    this.router.navigate(['/forgot-password/sent'], { state: { email: maskedEmail } });
  }

  async submit() {
    if (!this.validateEmail()) {
      return;
    }

    this.isSubmitting = true;
    const masked = this.maskEmail(this.email);

    try {
      await this.auth.requestPasswordReset(this.email);
    } catch (err) {
      console.error('password reset request failed', err);
      // Intentionally do not surface specific errors to avoid account enumeration
    } finally {
      this.isSubmitting = false;
      this.goToConfirmation(masked);
    }
  }
}
