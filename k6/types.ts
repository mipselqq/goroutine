export interface AuthHeader {
  [key: string]: string;
  Authorization: string;
}

export interface ColumnsCreateRaceSetup {
  authHeader: AuthHeader;
  boardId: string;
}

export interface TasksCreateRaceSetup {
  authHeader: AuthHeader;
  boardId: string;
  columnId: string;
}

export interface ColumnsDeleteCompactionSetup {
  authHeader: AuthHeader;
  boardId: string,
  testColumnIds: string[];
  lastColumnId: string;
}

export interface Column {
  id: string;
  position: number;
}

export interface Task {
  id: string;
  position: number;
}

export interface TasksDeleteCompactionSetup {
  authHeader: AuthHeader;
  boardId: string;
  columnId: string;
  testTaskIds: string[];
  lastTaskId: string;
}

export interface K6Response {
  status: number;
  body: string;
  json(selector: string): any;
}
