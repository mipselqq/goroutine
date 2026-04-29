export interface AuthHeader {
  [key: string]: string;
  Authorization: string;
}

export interface ColumnsRaceSetup {
  authHeader: AuthHeader;
  boardId: string;
}

export interface TasksRaceSetup {
  authHeader: AuthHeader;
  boardId: string;
  columnId: string;
}

export interface K6Response {
  status: number;
  body: string;
  json(selector: string): any;
}
