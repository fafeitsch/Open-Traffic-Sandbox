import {BrowserModule} from '@angular/platform-browser';
import {NgModule} from '@angular/core';

import {AppComponent} from './app.component';
import {MapViewModule} from './map-view/map-view.module';

@NgModule({
  declarations: [
    AppComponent,
  ],
  imports: [
    BrowserModule,
    MapViewModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule {
}
