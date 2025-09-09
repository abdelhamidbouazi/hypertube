import { CanActivateFn } from '@angular/router';
import { inject } from '@angular/core';
import { AuthService } from '~/app/auth/services/auth.service';

export const initGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  return !auth.isLoading();
};
