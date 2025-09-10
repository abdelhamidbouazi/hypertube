import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { ActivatedRoute } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-reset-password',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './reset_password.page.html',
})
export class ResetPasswordPage {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  password = '';
  email = this.route.snapshot.queryParamMap.get('email') ?? '';
  confirmPassword = '';
  errorMsg: string | null = null;
  confirmPasswordError: string | null = null;

  onPasswordInput() {
    if (this.password && this.password.length < 8) {
      this.confirmPasswordError = 'Password must be at least 8 characters';
    } else if (this.confirmPassword) {
      this.confirmPasswordError =
        this.password !== this.confirmPassword ? 'Passwords do not match' : null;
    } else {
      this.confirmPasswordError = null;
    }
  }

  onConfirmPasswordInput() {
    this.confirmPasswordError =
      this.password !== this.confirmPassword ? 'Passwords do not match' : null;
    if (!this.confirmPasswordError && this.password && this.password.length < 8) {
      this.confirmPasswordError = 'Password must be at least 8 characters';
    }
  }

  validateConfirmPassword() {
    this.onConfirmPasswordInput();
  }

  async submit() {
    this.errorMsg = null;
    this.confirmPasswordError = null;

    if (!this.password || this.password.length < 8) {
      this.confirmPasswordError = 'Password must be at least 8 characters';
      return;
    }
    if (this.password !== this.confirmPassword) {
      this.confirmPasswordError = 'Passwords do not match';
      return;
    }

    const token = this.route.snapshot.queryParamMap.get('token') ?? '';
    const email = this.route.snapshot.queryParamMap.get('email') ?? this.email;
    try {
      await this.auth.resetPassword(email, token, this.password);
      this.router.navigate(['/login']);
    } catch (e) {
      console.error(e);
      this.errorMsg = 'Failed to update password. Please try again.';
      setTimeout(() => {
        this.router.navigate(['/login']);
      }, 1000);
    }
  }
}
