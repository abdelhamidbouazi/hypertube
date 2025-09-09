import { Component } from '@angular/core';

@Component({
  standalone: true,
  selector: 'app-loading',
  template: `
    <div class="flex items-center justify-center h-screen w-screen bg-white">
      <div class="text-center">
        <svg
          class="animate-spin h-12 w-12 mx-auto text-gray-700"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
          ></path>
        </svg>
        <p class="mt-4 text-lg text-gray-700">Loading...</p>
      </div>
    </div>
  `,
})
export class LoadingComponent {}
