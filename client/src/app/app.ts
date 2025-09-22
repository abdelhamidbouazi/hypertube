import { Component, signal, OnInit, inject, computed } from '@angular/core';
import { RouterOutlet, Router } from '@angular/router';
import { AuthService } from './auth/services/auth.service';
import { CommonModule } from '@angular/common';
import { LoadingComponent } from './components/loading/loading.component';
import { NavbarComponent } from './components/navbar/navbar.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, CommonModule, LoadingComponent, NavbarComponent],
  templateUrl: './app.html',
})
export class App implements OnInit {
  protected readonly title = signal('hypertube');

  private auth = inject(AuthService);
  private router = inject(Router);

  get isLoading() {
    return this.auth.isLoading();
  }

  get showNavbar() {
    return this.auth.isAuthenticated() || this.router.url.startsWith('/discover')
    
  }

  ngOnInit() {
    if (this.isLoading) {
      this.router.navigate(['/loading']);
    }
  }
}
