// filepath: /src/app/pages/register/register.page.ts
import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';

@Component({
  standalone: true,
  selector: 'app-register',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './register.page.html',
})
export class RegisterPage {
  private auth = inject(AuthService);
  email = '';
  username = '';
  firstName = '';
  lastName = '';
  password = '';
  submit() {
    this.auth.register({
      email: this.email,
      username: this.username,
      firstName: this.firstName,
      lastName: this.lastName,
      password: this.password,
    });
  }
}
