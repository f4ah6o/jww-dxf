import DxfParser from "@f12o/dxf-parser";
import { Viewer } from "@f12o/three-dxf";
import "../styles.css";

declare const Go: new () => GoRuntime;

type WasmResult = {
  ok: boolean;
  data?: string;
  error?: string;
};

declare global {
  // Set by Go WASM runtime
  var jwwToDxf: ((data: Uint8Array) => WasmResult) | undefined;
  var jwwToDxfString: ((data: Uint8Array) => WasmResult) | undefined;
  var jwwGetVersion: (() => string) | undefined;
}

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void> | void;
}

type StatusType = "info" | "success" | "error";

type ProgressUpdate = {
  layerCount: number;
  entityCount: number;
  blockCount: number;
};

const BUILD_COMMIT = import.meta.env.VITE_COMMIT_HASH ?? "unknown";

const elements = {
  fileInput: document.getElementById("fileInput") as HTMLInputElement,
  fileLabel: document.getElementById("fileLabel") as HTMLSpanElement,
  convertBtn: document.getElementById("convertBtn") as HTMLButtonElement,
  status: document.getElementById("status") as HTMLDivElement,
  meta: document.getElementById("meta") as HTMLDListElement,
  dxfOutput: document.getElementById("dxfOutput") as HTMLTextAreaElement,
  downloadLink: document.getElementById("downloadLink") as HTMLAnchorElement,
  viewer: document.getElementById("viewer") as HTMLDivElement,
  viewerMessage: document.getElementById("viewerMessage") as HTMLDivElement,
  commitHash: document.getElementById("commitHash") as HTMLElement,
  jwwVersion: document.getElementById("jwwVersion") as HTMLElement,
};

let selectedFile: File | null = null;
let wasmReady = false;
let downloadUrl: string | null = null;

const go = new Go();
const wasmReadyPromise = loadWasm();

setCommitHash();
attachEventHandlers();

function attachEventHandlers(): void {
  elements.fileInput.addEventListener("change", (event) => {
    const target = event.target as HTMLInputElement;
    selectedFile = target.files?.[0] ?? null;
    elements.fileLabel.textContent = selectedFile?.name ?? "ファイルを選択";
    updateConvertButton();
  });

  elements.convertBtn.addEventListener("click", async () => {
    elements.convertBtn.disabled = true;
    await convertFile();
    updateConvertButton();
  });
}

function setStatus(message: string, type: StatusType = "info"): void {
  elements.status.textContent = message;
  elements.status.dataset.type = type;
}

function updateConvertButton(): void {
  elements.convertBtn.disabled = !wasmReady || !selectedFile;
}

async function loadWasm(): Promise<void> {
  setStatus("WASM を読み込んでいます…");
  try {
    const result = await WebAssembly.instantiateStreaming(
      fetch("/jww-parser.wasm"),
      go.importObject
    ).catch(async () => {
      const response = await fetch("/jww-parser.wasm");
      const bytes = await response.arrayBuffer();
      return WebAssembly.instantiate(bytes, go.importObject);
    });

    go.run(result.instance);
    wasmReady = true;
    setStatus("WASM がロードされました。JWW ファイルを選択してください。", "success");
    updateVersion();
    updateConvertButton();
  } catch (error) {
    console.error(error);
    const message =
      error instanceof Error ? error.message : "Unknown WASM load error";
    setStatus(`WASM のロードに失敗しました: ${message}`, "error");
  }
}

async function convertFile(): Promise<void> {
  if (!selectedFile) {
    alert("JWW ファイルを選択してください。");
    return;
  }

  await wasmReadyPromise;
  if (!wasmReady) {
    alert("WASM のロードに失敗しています。ページを再読み込みしてください。");
    return;
  }

  try {
    setStatus("JWW を読み込み中…");
    const buffer = await selectedFile.arrayBuffer();
    const bytes = new Uint8Array(buffer);

    const dxfDocResult = globalThis.jwwToDxf?.(bytes);
    if (!dxfDocResult?.ok || !dxfDocResult.data) {
      throw new Error(dxfDocResult?.error || "DXF 変換に失敗しました");
    }

    const doc = JSON.parse(dxfDocResult.data) as ProgressUpdate & {
      Layers?: unknown[];
      Entities?: unknown[];
      Blocks?: unknown[];
    };
    updateMeta({
      layerCount: doc?.Layers?.length ?? 0,
      entityCount: doc?.Entities?.length ?? 0,
      blockCount: doc?.Blocks?.length ?? 0,
    });

    setStatus("DXF を生成しています…");
    const dxfStringResult = globalThis.jwwToDxfString?.(bytes);
    if (!dxfStringResult?.ok || !dxfStringResult.data) {
      throw new Error(dxfStringResult?.error || "DXF の生成に失敗しました");
    }

    const dxfString = dxfStringResult.data;
    elements.dxfOutput.value = dxfString;
    updateDownloadLink(dxfString);
    renderPreview(dxfString);
    setStatus(
      "DXF を生成しました。プレビューとダウンロードが利用できます。",
      "success"
    );
  } catch (error) {
    console.error(error);
    const message = error instanceof Error ? error.message : String(error);
    setStatus(`変換に失敗しました: ${message}`, "error");
    resetViewer("プレビューを表示できませんでした。");
    elements.downloadLink.classList.add("disabled");
    elements.downloadLink.setAttribute("aria-disabled", "true");
    elements.downloadLink.removeAttribute("href");
  }
}

function updateMeta(update: ProgressUpdate): void {
  const entries = [
    { label: "レイヤー数", value: update.layerCount },
    { label: "エンティティ数", value: update.entityCount },
    { label: "ブロック数", value: update.blockCount },
  ];

  elements.meta.innerHTML = entries
    .map(
      (entry) => `
        <div>
            <dt>${entry.label}</dt>
            <dd>${entry.value}</dd>
        </div>
      `
    )
    .join("");
}

function updateDownloadLink(dxfString: string): void {
  if (downloadUrl) {
    URL.revokeObjectURL(downloadUrl);
  }

  const blob = new Blob([dxfString], { type: "application/dxf" });
  downloadUrl = URL.createObjectURL(blob);

  elements.downloadLink.href = downloadUrl;
  elements.downloadLink.download = `${selectedFile?.name || "output"}.dxf`;
  elements.downloadLink.classList.remove("disabled");
  elements.downloadLink.setAttribute("aria-disabled", "false");
}

function renderPreview(dxfString: string): void {
  resetViewer("プレビューを準備しています…");

  try {
    const parser = new DxfParser();
    const parsed = parser.parseSync(dxfString);
    const width = elements.viewer.clientWidth || 640;
    const height = elements.viewer.clientHeight || 480;

    const viewer = new Viewer(parsed, elements.viewer, width, height);
    viewer.renderer?.setClearColor?.(0x000000, 1);
    viewer.render();
    elements.viewerMessage.textContent =
      "右クリックでパン、マウスホイールでズームできます。";
  } catch (error) {
    console.error(error);
    const message = error instanceof Error ? error.message : String(error);
    resetViewer(`DXF プレビューの生成に失敗しました: ${message}`);
  }
}

function resetViewer(message: string): void {
  elements.viewer.innerHTML = "";
  elements.viewer.appendChild(elements.viewerMessage);
  elements.viewerMessage.textContent = message;
}

function setCommitHash(): void {
  if (!elements.commitHash) return;

  elements.commitHash.textContent = BUILD_COMMIT;
}

function updateVersion(): void {
  if (!elements.jwwVersion || typeof globalThis.jwwGetVersion !== "function")
    return;

  try {
    const version = globalThis.jwwGetVersion();
    elements.jwwVersion.textContent = version || "unknown";
  } catch (error) {
    console.error(error);
    elements.jwwVersion.textContent = "unknown";
  }
}
