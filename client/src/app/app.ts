import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
<<<<<<< HEAD
  protected readonly title = signal('hypertybe-learning');
=======
  protected readonly title = signal('client');
>>>>>>> 669195e367f919c154b0c3d6f114408079d8fd77
}
