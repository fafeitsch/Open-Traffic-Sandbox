import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {LineServiceModule} from '../line-service/line-service.module';
import {LineService} from '../line-service/line.service';


@NgModule({
  declarations: [],
  imports: [
    CommonModule, LineServiceModule
  ],
  providers: [LineService]
})
export class ServiceModule {
}
