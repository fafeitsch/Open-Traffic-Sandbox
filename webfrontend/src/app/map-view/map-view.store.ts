import {Injectable} from '@angular/core';
import {ComponentStore} from '@ngrx/component-store';
import {tap} from 'rxjs/operators';
import {Line, LineRoute} from '../line-service/types';
import {LineViewService} from '../service/line-view.service';

interface State {
  visibleLines: { [key: string]: LineRoute };
}

@Injectable()
export class MapViewStore extends ComponentStore<State> {
  constructor(private lineViewService: LineViewService) {
    super({visibleLines: {}});
  }

  readonly visibleLines$ = this.select(state => state.visibleLines);

  readonly addLineRoute$ = this.updater((state, route: LineRoute) => {
      const lines = {...state.visibleLines};
      lines[route.key] = route;
      return {...state, visibleLines: lines};
    }
  );

  readonly removeLine$ = this.updater((state, line: Line) => {
      const lines = {...state.visibleLines};
      delete lines[line.key];
      return {...state, visibleLines: lines};
    }
  );

  readonly initAddLineListener$ = this.effect(() =>
    this.lineViewService.activateLineSubject.pipe(
      tap<LineRoute>(route => this.addLineRoute$(route))
    )
  );

  readonly initRemoveLineListener$ = this.effect(() =>
    this.lineViewService.deactivateLineSubject.pipe(
      tap<Line>(line => this.removeLine$(line))
    )
  );
}
