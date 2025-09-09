import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '~/app/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-browse',
  imports: [CommonModule],
  templateUrl: './browse.page.html',
})
export class BrowsePage implements OnInit {
  private auth = inject(AuthService);
  user: any = null;

  ngOnInit() {
    this.loadUserData();
  }

  logout() {
    this.auth.logout();
  }

  getMe() {
    this.loadUserData();
  }

  private async loadUserData() {
    try {
      await this.auth.loadMe();
      // After loadMe completes, get user from the auth service
      this.user = this.auth.user();
      console.log('Loaded 1 user data:', this.user);

      // Debug the actual property names and values
      if (this.user) {
        console.log('User object properties:', Object.keys(this.user));
        console.log('User object is actually set:', !!this.user);
      }
    } catch (error) {
      console.error('Error loading user data:', error);
    }
  }
}
