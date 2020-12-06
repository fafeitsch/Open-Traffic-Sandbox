import {Injectable} from '@angular/core';
import {ComponentStore} from '@ngrx/component-store';
import {Observable} from 'rxjs';
import {switchMap, tap} from 'rxjs/operators';
import {LineService} from '../line-service/line.service';
import {Line, LineRoute} from '../line-service/types';
import {LineViewService} from '../service/line-view.service';

interface LineTableState {
  lines: Line[];
}

@Injectable()
export class LineListStore extends ComponentStore<LineTableState> {
  constructor(private service: LineService,
              private lineViewService: LineViewService) {
    super({lines: []});
  }

  readonly lines$ = super.select(state => state.lines);

  readonly setLines$ = super.updater((state, lines: Line[]) => ({...state, lines: lines}));

  readonly loadLines$ = super.effect(() => this.service.getLines().pipe(tap<Line[]>(lines => this.setLines$(lines))));

  readonly viewLine$ = super.effect((line$: Observable<Line>) =>
    line$.pipe(
      switchMap(line => this.service.getLineRoute(line.id)),
      tap<LineRoute>(route => this.lineViewService.activateLine(route))
    )
  );

  readonly removeLine$ = super.effect((line$: Observable<Line>) =>
    line$.pipe(
      tap<Line>(line => this.lineViewService.deactivateLine(line))
    )
  );
}
