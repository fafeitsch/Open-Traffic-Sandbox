import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MapViewComponent} from './map-view.component';
import {ClipboardModule} from '@angular/cdk/clipboard';


@NgModule({
  declarations: [MapViewComponent],
  exports: [MapViewComponent],
  imports: [
    CommonModule, ClipboardModule
  ]
})
export class MapViewModule {
}
