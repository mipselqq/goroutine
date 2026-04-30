import http from "k6/http";
import type { Options } from "k6/options";
import { createBoard, createColumn, createTaskRequest, defaultRegisterAndLogin, deleteBoard } from "./prelude.ts";
import type { TasksCreateRaceSetup } from "./types.ts";
import { check } from "k6";

export function setup(): TasksCreateRaceSetup {
    const authHeader = defaultRegisterAndLogin();
    const boardId = createBoard(authHeader);
    const columnId = createColumn(boardId, authHeader);
    return { authHeader, boardId, columnId };
}

export const options: Options = {
    scenarios: {
        tasksRace: {
            executor: "per-vu-iterations",
            vus: 1,
            iterations: 1,
        },
    },
    thresholds: {
        checks: ["rate == 1"],
    },
};

// Attempting to catch 500 Internal Error because of Unique Violation
export default function tasksRace({ authHeader, boardId, columnId }: TasksCreateRaceSetup): void {
    const batch = http.batch(
        Array.from({ length: 10 }, () => createTaskRequest(boardId, columnId, authHeader)),
    );

    for (const response of batch) {
        check(response, { "201": (x) => x.status === 201 });
    }
}

export function teardown({ authHeader, boardId }: TasksCreateRaceSetup): void {
    deleteBoard(boardId, authHeader);
}
