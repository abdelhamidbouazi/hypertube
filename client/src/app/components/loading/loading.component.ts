import { Component } from '@angular/core';
import { NzSpinModule } from 'ng-zorro-antd/spin';

@Component({
  standalone: true,
  selector: 'app-loading',
  imports: [NzSpinModule],
  template: `
    <div class="flex items-center justify-center h-screen w-screen bg-white">
      <nz-spin  nzSize="large"></nz-spin>
    </div>
  `,
})
export class LoadingComponent {}
