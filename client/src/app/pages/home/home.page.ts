import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NzIconModule, NzIconService } from 'ng-zorro-antd/icon';
import { ArrowRightOutline } from '@ant-design/icons-angular/icons';

@Component({
  standalone: true,
  selector: 'app-home',
  imports: [RouterLink, NzIconModule],
  templateUrl: './home.page.html',
})
export class HomePage {
  constructor(private iconService: NzIconService) {
    this.iconService.addIcon(ArrowRightOutline);
  }
}
