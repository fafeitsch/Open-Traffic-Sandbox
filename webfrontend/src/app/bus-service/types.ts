import { Line } from '../line-service/types';

export interface BusInfo {
  name: string;
  id: string;
  assignment: string;
  line: Line;
}
