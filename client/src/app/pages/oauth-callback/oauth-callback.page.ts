import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  standalone: true,
  selector: 'app-oauth-callback',
  imports: [CommonModule],
  templateUrl: './oauth-callback.page.html',
})
export class OAuthCallbackPage {}
