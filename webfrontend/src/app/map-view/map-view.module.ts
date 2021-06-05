import { ClipboardModule } from '@angular/cdk/clipboard';
import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { BusServiceModule } from '../bus-service/bus-service.module';
import { ServiceModule } from '../service/service.module';
import { MapViewComponent } from './map-view.component';

@NgModule({
  declarations: [MapViewComponent],
  exports: [MapViewComponent],
  imports: [CommonModule, ClipboardModule, ServiceModule, BusServiceModule],
})
export class MapViewModule {}
