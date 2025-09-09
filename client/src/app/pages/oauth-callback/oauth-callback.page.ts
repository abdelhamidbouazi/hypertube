import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-oauth-callback',
  imports: [CommonModule],
  templateUrl: './oauth-callback.page.html',
})
export class OAuthCallbackPage implements OnInit {
  private router = inject(Router);
  private auth = inject(AuthService);

  async ngOnInit() {
    try {
      await this.auth.completeOAuthFromCookies();
    } catch (err) {
      console.error('OAuth login failed', err);
      this.router.navigate(['/login']);
    }
  }
}
