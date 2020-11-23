import {Component, OnInit} from '@angular/core';
import {LineListStore} from './line-list-store.service';
import {Line} from '../line-service/types';

@Component({
  selector: 'line-list',
  templateUrl: './line-list.component.html',
  styleUrls: ['./line-list.component.scss'],
  providers: [LineListStore]
})
export class LineListComponent implements OnInit {

  lines$ = this.store.lines$;

  constructor(private store: LineListStore) {
  }

  ngOnInit(): void {
  }

  activateLine(line: Line) {
    this.store.viewLine$(line);
  }

  deactivateLine(line: Line) {
    this.store.removeLine$(line);
  }
}
