import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';

@Component({
  standalone: true,
  selector: 'app-forgot-password-sent',
  imports: [CommonModule, RouterLink],
  templateUrl: './forgot-password-sent.page.html',
})
export class ForgotPasswordSentPage {
  private router = inject(Router);

  get email(): string | null {
    const nav = this.router.getCurrentNavigation();
    const state = nav?.extras?.state as { email?: string } | undefined;
    return state?.email ?? null;
  }

  openEmailApp() {
    // Best-effort: open default mail client
    window.location.href = 'mailto:';
  }
}
