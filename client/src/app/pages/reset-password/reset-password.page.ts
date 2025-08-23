import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ActivatedRoute } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-reset-password',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './reset-password.page.html',
})
export class ResetPasswordPage {
  private route = inject(ActivatedRoute);
  private auth = inject(AuthService);
  password = '';
  confirmPassword = '';

  submit() {
    if (this.password !== this.confirmPassword) {
      alert('Passwords do not match');
      return;
    }
    const token = this.route.snapshot.queryParamMap.get('token') ?? '';
    this.auth.resetPassword(token, this.password);
  }
}
