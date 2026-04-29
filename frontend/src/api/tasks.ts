import { fetchURL, fetchJSON } from "./utils";

export function taskDelete(taskID: string) {
  const url = `?id=${taskID}`;
  return tasksFetch(url, "DELETE");
}

export function taskList() {
  const url = "/api/tasks/";
  const method = "GET";
  return fetchJSON(url, { method });
}

export function taskCall(data: { [key: string]: string }) {
  const { taskName, ...body } = data;
  const url = `?taskName=${taskName}`;
  return tasksFetch(url, "POST", body);
}

async function tasksFetch(url: string, method: ApiMethod, content?: any) {
  // url = removePrefix(url);

  const opts: ApiOpts = {
    method,
  };

  if (content) {
    opts.body = JSON.stringify(content);
  }

  const res = await fetchURL(`/api/tasks/${url}`, opts);

  return res;
}
