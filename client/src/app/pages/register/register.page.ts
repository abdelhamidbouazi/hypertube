import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';
import { MatIconModule } from '@angular/material/icon';

@Component({
  standalone: true,
  selector: 'app-register',
  imports: [CommonModule, FormsModule, RouterLink, MatIconModule],
  templateUrl: './register.page.html',
})
export class RegisterPage {
  private auth = inject(AuthService);
  email = '';
  firstName = '';
  lastName = '';
  password = '';
  isLoading = false;

  emailError = '';
  firstNameError = '';
  lastNameError = '';
  passwordError = '';
  formError = '';

  readonly MIN_PASSWORD_LENGTH = 8;

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

  validateFirstName(): boolean {
    this.firstName = this.firstName.trim();
    if (!this.firstName) {
      this.firstNameError = 'First name is required';
      return false;
    }

    this.firstNameError = '';
    return true;
  }

  validateLastName(): boolean {
    this.lastName = this.lastName.trim();
    if (!this.lastName) {
      this.lastNameError = 'Last name is required';
      return false;
    }

    this.lastNameError = '';
    return true;
  }

  validatePassword(): boolean {
    if (!this.password) {
      this.passwordError = 'Password is required';
      return false;
    }

    if (this.password.length < this.MIN_PASSWORD_LENGTH) {
      this.passwordError = `Password must be at least ${this.MIN_PASSWORD_LENGTH} characters long`;
      return false;
    }

    const hasUpperCase = /[A-Z]/.test(this.password);
    const hasLowerCase = /[a-z]/.test(this.password);
    const hasNumbers = /\d/.test(this.password);
    // const hasSpecialChar = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+/.test(this.password);

    if (!(hasUpperCase && hasLowerCase && hasNumbers)) {
      this.passwordError = 'Password must contain uppercase, lowercase letters and numbers';
      return false;
    }

    this.passwordError = '';
    return true;
  }

  validateForm(): boolean {
    const isEmailValid = this.validateEmail();
    const isFirstNameValid = this.validateFirstName();
    const isLastNameValid = this.validateLastName();
    const isPasswordValid = this.validatePassword();

    return isEmailValid && isFirstNameValid && isLastNameValid && isPasswordValid;
  }

  async submit() {
    if (this.isLoading) return;

    this.formError = '';

    if (!this.validateForm()) {
      return;
    }

    this.isLoading = true;
    try {
      await this.auth.register({
        email: this.email,
        firstName: this.firstName,
        lastName: this.lastName,
        password: this.password,
      });
    } catch (err) {
      console.error('register failed', err);
      this.formError = 'Registration failed. Please try again.';
    } finally {
      this.isLoading = false;
    }
  }

  onEmailInput() {
    this.emailError = '';
  }

  onFirstNameInput() {
    this.firstNameError = '';
  }

  onLastNameInput() {
    this.lastNameError = '';
  }

  onPasswordInput() {
    this.passwordError = '';
  }
}
