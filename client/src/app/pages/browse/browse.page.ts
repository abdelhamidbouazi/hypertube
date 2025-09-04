import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AuthService } from '~/app/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-browse',
  imports: [CommonModule],
  templateUrl: './browse.page.html',
})
export class BrowsePage {
  private authService = inject(AuthService);

  logout() {
    this.authService.logout();
  }
}
