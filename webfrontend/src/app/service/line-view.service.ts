import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Line, LineRoute } from '../line-service/types';
import { ServiceModule } from './service.module';

@Injectable({ providedIn: ServiceModule })
export class LineViewService {
  private readonly _activateLineSubject = new Subject<LineRoute>();
  private _deactivateLineSubject = new Subject<Line>();

  constructor() {
    this._activateLineSubject = new Subject<LineRoute>();
  }

  get activateLineSubject(): Subject<LineRoute> {
    return this._activateLineSubject;
  }

  get deactivateLineSubject(): Subject<Line> {
    return this._deactivateLineSubject;
  }

  activateLine(route: LineRoute) {
    this.activateLineSubject.next(route);
  }

  deactivateLine(line: Line) {
    this._deactivateLineSubject.next(line);
  }
}
