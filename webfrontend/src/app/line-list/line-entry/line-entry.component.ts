import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Line} from '../../line-service/types';

@Component({
  selector: 'line-entry',
  templateUrl: './line-entry.component.html',
  styleUrls: ['./line-entry.component.scss']
})
export class LineEntryComponent {

  visible = false;

  @Input() line: Line;
  @Output() makeVisible = new EventEmitter<Line>();
  @Output() makeInvisible = new EventEmitter<Line>();

  constructor() {
  }

  toggleVisibility() {
    if (this.visible) {
      this.makeInvisible.emit(this.line);
    } else {
      this.makeVisible.emit(this.line);
    }
    this.visible = !this.visible;
  }
}
