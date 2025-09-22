import { Component, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '@/auth/services/auth.service';
import { NzImageModule } from 'ng-zorro-antd/image';
import { NzAffixModule } from 'ng-zorro-antd/affix';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzBadgeModule } from 'ng-zorro-antd/badge';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    NzImageModule,
    NzAffixModule,
    NzLayoutModule,
    NzMenuModule,
    NzButtonModule,
    NzIconModule,
    NzBadgeModule,
    NzAvatarModule,
  ],
  template: `
    <nz-affix [nzOffsetTop]="0">
      <nz-layout class="bg-slate-900">
        <nz-header class="z-50 bg-slate-900 p-0">
          <div class="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
            <div class="h-16 flex items-center justify-between">
              <!-- Left: Brand -->
              <a
                [routerLink]="isAuthed() ? '/discover' : '/login'"
                class="flex items-center gap-2 select-none no-underline"
              >
                <img
                  nz-image
                  width="150px"
                  nzSrc="/assets/logo-white.svg"
                  alt="cinethos logo"
                />
              </a>

              <!-- Center: Navigation -->
              <ul
                nz-menu
                nzMode="horizontal"
                nzTheme="dark"
                class="hidden md:flex items-center bg-transparent border-0"
              >
                <li nz-menu-item [routerLink]="'/discover'" [nzSelected]="isOnDiscover()">Home</li>
                <li nz-menu-item [routerLink]="'/categories'" >
                  Categories
                </li>
                <li nz-menu-item [routerLink]="'/settings'" >
                  Settings
                </li>
              </ul>

              <!-- Right: Actions -->
              <div class="flex items-center gap-3 text-white">
                <button nz-button nzType="text" nzShape="circle" aria-label="Search">
                  <span nz-icon nzType="search" class="text-white"></span>
                </button>
                
                <a [routerLink]="isAuthed() ? '/browse' : '/login'" aria-label="Account">
                  <nz-avatar nzSize="small" nzIcon="user"></nz-avatar>
                </a>
              </div>
            </div>
          </div>
          <div class="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>
        </nz-header>
      </nz-layout>
    </nz-affix>
  `,
})
export class NavbarComponent {
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  protected readonly isAuthed = computed(() => this.auth.isAuthenticated());

  protected isOnDiscover(): boolean {
    return this.router.url.startsWith('/discover');
  }

  protected isOnBrowse(): boolean {
    return this.router.url.startsWith('/browse');
  }
}
