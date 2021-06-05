import { ChangeDetectionStrategy, Component } from '@angular/core';

@Component({
  selector: 'main-view',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AppComponent {
  title = 'webfrontend';
}
