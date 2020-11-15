import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MapViewComponent} from './map-view.component';


@NgModule({
  declarations: [MapViewComponent],
  exports: [MapViewComponent],
  imports: [
    CommonModule
  ]
})
export class MapViewModule {
}
