import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MapViewComponent} from './map-view.component';
import {ClipboardModule} from '@angular/cdk/clipboard';
import {ServiceModule} from '../service/service.module';


@NgModule({
  declarations: [MapViewComponent],
  exports: [MapViewComponent],
  imports: [
    CommonModule, ClipboardModule, ServiceModule
  ]
})
export class MapViewModule {
}
