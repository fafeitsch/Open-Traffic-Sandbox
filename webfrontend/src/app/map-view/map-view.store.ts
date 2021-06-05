import { Injectable } from '@angular/core';
import { ComponentStore } from '@ngrx/component-store';
import { Observable } from 'rxjs';
import { filter, switchMap, tap } from 'rxjs/operators';
import { BusService } from '../bus-service/bus.service';
import { BusInfo } from '../bus-service/types';
import { Line, LineRoute } from '../line-service/types';
import { LineViewService } from '../service/line-view.service';

interface State {
  visibleLines: { [key: string]: LineRoute };
  currentBus: BusInfo | undefined;
}

@Injectable()
export class MapViewStore extends ComponentStore<State> {
  readonly visibleLines$ = this.select(state => state.visibleLines);
  readonly currentBus$ = this.select(state => state.currentBus).pipe(filter(info => info !== undefined));

  readonly addLineRoute$ = this.updater((state, route: LineRoute) => {
    const lines = { ...state.visibleLines };
    lines[route.key] = route;
    return { ...state, visibleLines: lines };
  });

  readonly removeLine$ = this.updater((state, line: Line) => {
    const lines = { ...state.visibleLines };
    delete lines[line.id];
    return { ...state, visibleLines: lines };
  });

  readonly setCurrentBus$ = this.updater((state, bus: BusInfo) => ({ ...state, currentBus: bus }));

  readonly initAddLineListener$ = this.effect(() =>
    this.lineViewService.activateLineSubject.pipe(tap<LineRoute>(route => this.addLineRoute$(route)))
  );

  readonly initRemoveLineListener$ = this.effect(() =>
    this.lineViewService.deactivateLineSubject.pipe(tap<Line>(line => this.removeLine$(line)))
  );

  readonly busSelectionChanged$ = this.effect((id$: Observable<string>) =>
    id$.pipe(
      switchMap(id => this.bussService.getBusInfo(id)),
      tap<BusInfo>(info => this.setCurrentBus$(info))
    )
  );

  constructor(private lineViewService: LineViewService, private bussService: BusService) {
    super({ visibleLines: {}, currentBus: undefined });
  }
}
