import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { LineServiceModule } from '../line-service/line-service.module';
import { ServiceModule } from '../service/service.module';
import { LineListComponent } from './line-list.component';
import { LineEntryComponent } from './line-entry/line-entry.component';

@NgModule({
  declarations: [LineListComponent, LineEntryComponent],
  exports: [LineListComponent],
  imports: [CommonModule, LineServiceModule, MatIconModule, MatButtonModule, ServiceModule],
})
export class LineListModule {}
