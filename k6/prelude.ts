import http from "k6/http";
import { check } from "k6";
import type { AuthHeader, Column, Task } from "./types.ts";

export const API_BASE = __ENV.K6_ROOT || "http://localhost:8080";
export const PWD = "testPassword$123";
export const JSON_HEADER = { "Content-Type": "application/json" };

export function generateUniqueEmail(): string {
    return `vu${__VU}-${Date.now()}@t.t`;
}

export function createBoard(authHeader: AuthHeader): string {
    const boardResp = http.post(
        `${API_BASE}/v1/boards`,
        JSON.stringify({ name: "Test board", description: "" }),
        { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "createBoard" } },
    );
    check(boardResp, { "createBoard status is 201": (r) => r.status === 201 });
    if (boardResp.status !== 201) {
        throw new Error(`create board failed: ${boardResp.status} ${boardResp.body}`);
    }

    return boardResp.json("id") as string;
}

export function createColumnRequest(boardId: string, authHeader: AuthHeader) {
    return {
        method: "POST",
        url: `${API_BASE}/v1/boards/${boardId}/columns`,
        params: { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "createColumn" } },
        body: JSON.stringify({ name: "Test column", description: "" }),
    };
}

export function createColumn(boardId: string, authHeader: AuthHeader): string {
    const req = createColumnRequest(boardId, authHeader);

    const columnResp = http.request(req.method, req.url, req.body, req.params);
    check(columnResp, { "createColumn status is 201": (r) => r.status === 201 });
    if (columnResp.status !== 201) {
        throw new Error(`create column failed: ${columnResp.status} ${columnResp.body}`);
    }

    return columnResp.json("id") as string;
}

export function listColumns(boardId: string, authHeader: AuthHeader): Column[] {
    const listResp = http.get(
        `${API_BASE}/v1/boards/${boardId}/columns`,
        { headers: authHeader, tags: { name: "listColumns" } },
    );
    check(listResp, { "listColumns status is 200": (r) => r.status === 200 });
    if (listResp.status !== 200) {
        throw new Error(`list columns failed: ${listResp.status} ${listResp.body}`);
    }
    return listResp.json() as unknown as Column[];
}

export function deleteColumn(boardId: string, columnId: string, authHeader: AuthHeader): void {
    const delResp = http.del(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}`,
        null,
        { headers: authHeader, tags: { name: "deleteColumn" } },
    );
    check(delResp, { "deleteColumn status is 204": (r) => r.status === 204 });
    if (delResp.status !== 204) {
        throw new Error(`delete column failed: ${delResp.status} ${delResp.body}`);
    }
}

export function deleteColumnRequest(boardId: string, columnId: string, authHeader: AuthHeader) {
    return {
        method: "DELETE",
        url: `${API_BASE}/v1/boards/${boardId}/columns/${columnId}`,
        params: { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "deleteColumn" } },
    };
}

export function getBoard(boardId: string, authHeader: AuthHeader): void {
    const boardResp = http.get(
        `${API_BASE}/v1/boards/${boardId}`,
        { headers: authHeader, tags: { name: "getBoard" } },
    );
    check(boardResp, { "getBoard status is 200": (r) => r.status === 200 });
    if (boardResp.status !== 200) {
        throw new Error(`get board failed: ${boardResp.status} ${boardResp.body}`);
    }
}

export function listBoards(authHeader: AuthHeader): void {
    const listResp = http.get(
        `${API_BASE}/v1/boards`,
        { headers: authHeader, tags: { name: "listBoards" } },
    );
    check(listResp, { "listBoards status is 200": (r) => r.status === 200 });
    if (listResp.status !== 200) {
        throw new Error(`list boards failed: ${listResp.status} ${listResp.body}`);
    }
}

export function getColumn(boardId: string, columnId: string, authHeader: AuthHeader): Column {
    const listResp = http.get(
        `${API_BASE}/v1/boards/${boardId}/columns`,
        { headers: authHeader, tags: { name: "getColumn" } },
    );
    check(listResp, { "getColumn status is 200": (r) => r.status === 200 });
    if (listResp.status !== 200) {
        throw new Error(`list columns failed: ${listResp.status} ${listResp.body}`);
    }

    const column = (listResp.json() as unknown as Column[]).find((c) => c.id === columnId);
    if (!column) {
        throw new Error(`column ${columnId} not found in list`);
    }
    return column;
}

export function deleteBoard(boardId: string, authHeader: AuthHeader): void {
    const deleteResp = http.del(
        `${API_BASE}/v1/boards/${boardId}`,
        null,
        { headers: authHeader, tags: { name: "deleteBoard" } },
    );
    check(deleteResp, { "deleteBoard status is 204": (r) => r.status === 204 });
    if (deleteResp.status !== 204) {
        throw new Error(`delete board failed: ${deleteResp.status} ${deleteResp.body}`);
    }
}

export function listTasks(boardId: string, columnId: string, authHeader: AuthHeader): Task[] {
    const listResp = http.get(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks`,
        { headers: authHeader, tags: { name: "listTasks" } },
    );
    check(listResp, { "listTasks status is 200": (r) => r.status === 200 });
    if (listResp.status !== 200) {
        throw new Error(`list tasks failed: ${listResp.status} ${listResp.body}`);
    }
    return listResp.json() as unknown as Task[];
}

export function moveTask(
    boardId: string,
    columnId: string,
    taskId: string,
    targetColumnId: string,
    targetPosition: number,
    authHeader: AuthHeader,
): void {
    const moveResp = http.put(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks/${taskId}/position`,
        JSON.stringify({ targetColumnId, targetPosition }),
        { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "moveTask" } },
    );
    check(moveResp, { "moveTask status is 200": (r) => r.status === 200 });
    if (moveResp.status !== 200) {
        throw new Error(`move task failed: ${moveResp.status} ${moveResp.body}`);
    }
}

export function deleteTask(boardId: string, columnId: string, taskId: string, authHeader: AuthHeader): void {
    const delResp = http.del(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks/${taskId}`,
        null,
        { headers: authHeader, tags: { name: "deleteTask" } },
    );
    check(delResp, { "deleteTask status is 204": (r) => r.status === 204 });
    if (delResp.status !== 204) {
        throw new Error(`delete task failed: ${delResp.status} ${delResp.body}`);
    }
}

export function getAggregate(boardId: string, authHeader: AuthHeader): void {
    const aggResp = http.get(
        `${API_BASE}/v1/boards/${boardId}/aggregate`,
        { headers: authHeader, tags: { name: "getAggregate" } },
    );
    check(aggResp, { "getAggregate status is 200": (r) => r.status === 200 });
    if (aggResp.status !== 200) {
        throw new Error(`get aggregate failed: ${aggResp.status} ${aggResp.body}`);
    }
}

export function updateTask(
    boardId: string,
    columnId: string,
    taskId: string,
    name: string,
    description: string,
    authHeader: AuthHeader,
): void {
    const patchResp = http.patch(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks/${taskId}`,
        JSON.stringify({ name, description }),
        { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "updateTask" } },
    );
    check(patchResp, { "updateTask status is 200": (r) => r.status === 200 });
    if (patchResp.status !== 200) {
        throw new Error(`update task failed: ${patchResp.status} ${patchResp.body}`);
    }
}

export function createTaskRequest(boardId: string, columnId: string, authHeader: AuthHeader) {
    return {
        method: "POST",
        url: `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks`,
        params: { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "createTask" } },
        body: JSON.stringify({ name: "Test task", description: "" }),
    };
}

export function createTask(boardId: string, columnId: string, authHeader: AuthHeader): string {
    const req = createTaskRequest(boardId, columnId, authHeader);

    const taskResp = http.request(req.method, req.url, req.body, req.params);
    check(taskResp, { "createTask status is 201": (r) => r.status === 201 });
    if (taskResp.status !== 201) {
        throw new Error(`create task failed: ${taskResp.status} ${taskResp.body}`);
    }
    return taskResp.json("id") as string;
}

export function deleteTaskRequest(boardId: string, columnId: string, taskId: string, authHeader: AuthHeader) {
    return {
        method: "DELETE",
        url: `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks/${taskId}`,
        params: { headers: { ...JSON_HEADER, ...authHeader }, tags: { name: "deleteTask" } },
    };
}

export function getTask(boardId: string, columnId: string, taskId: string, authHeader: AuthHeader): Task {
    const listResp = http.get(
        `${API_BASE}/v1/boards/${boardId}/columns/${columnId}/tasks`,
        { headers: authHeader, tags: { name: "getTask" } },
    );
    check(listResp, { "getTask status is 200": (r) => r.status === 200 });
    if (listResp.status !== 200) {
        throw new Error(`list tasks failed: ${listResp.status} ${listResp.body}`);
    }

    const task = (listResp.json() as unknown as Task[]).find((t) => t.id === taskId);
    if (!task) {
        throw new Error(`task ${taskId} not found in list`);
    }
    return task;
}

export function defaultRegisterAndLogin(): AuthHeader {
    const email = generateUniqueEmail();
    const password = PWD;

    const registerResp = http.post(
        `${API_BASE}/v1/register`,
        JSON.stringify({ email, password }),
        { headers: { "Content-Type": "application/json" }, tags: { name: "register" } },
    );
    check(registerResp, { "register status is 200": (r) => r.status === 200 });
    if (registerResp.status !== 200) {
        throw new Error(`register failed: ${registerResp.status} ${registerResp.body}`);
    }

    const loginResp = http.post(
        `${API_BASE}/v1/login`,
        JSON.stringify({ email, password }),
        { headers: { "Content-Type": "application/json" }, tags: { name: "login" } },
    );
    check(loginResp, { "login status is 200": (r) => r.status === 200 });
    if (loginResp.status !== 200) {
        throw new Error(`login failed: ${loginResp.status} ${loginResp.body}`);
    }

    const token = loginResp.json("token");

    return { Authorization: `Bearer ${token}` };
}
