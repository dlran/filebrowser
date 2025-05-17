import { fetchURL, removePrefix } from "./utils";

export function copyExif(item: {[key: string]: string}) {
  const from = item.from;
  const to = encodeURIComponent(removePrefix(item.to ?? ""));
  const url = `${from}?action=copyExif&destination=${to}&fps=true&rename=false`;
  return toolsFetch(url, "PATCH");
}

export function extractFrame(item: {[key: string]: string}) {
  const from = item.from;
  const url = `${from}?action=extractFrame&fps=${item.fps}`;
  return toolsFetch(url, "PATCH");
}

async function toolsFetch(url: string, method: ApiMethod, content?: any) {
  url = removePrefix(url);

  const opts: ApiOpts = {
    method,
  };

  if (content) {
    opts.body = content;
  }

  const res = await fetchURL(`/api/tools${url}`, opts);

  return res;
}
