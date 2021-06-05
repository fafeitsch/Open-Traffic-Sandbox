import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppComponent } from './app.component';
import { MapViewModule } from './map-view/map-view.module';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { LineListModule } from './line-list/line-list.module';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatTabsModule } from '@angular/material/tabs';
import { MatIconModule } from '@angular/material/icon';

@NgModule({
  declarations: [AppComponent],
  imports: [
    BrowserModule,
    MapViewModule,
    BrowserAnimationsModule,
    LineListModule,
    MatSidenavModule,
    MatTabsModule,
    MatIconModule,
  ],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
