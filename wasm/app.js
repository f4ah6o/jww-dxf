(() => {
    const elements = {
        fileInput: document.getElementById("fileInput"),
        fileLabel: document.getElementById("fileLabel"),
        convertBtn: document.getElementById("convertBtn"),
        status: document.getElementById("status"),
        meta: document.getElementById("meta"),
        dxfOutput: document.getElementById("dxfOutput"),
        downloadLink: document.getElementById("downloadLink"),
        viewer: document.getElementById("viewer"),
        viewerMessage: document.getElementById("viewerMessage"),
    };

    let selectedFile = null;
    let wasmReady = false;
    let downloadUrl = null;

    const go = new Go();
    const wasmReadyPromise = loadWasm();

    elements.fileInput.addEventListener("change", (event) => {
        selectedFile = event.target.files?.[0] || null;
        elements.fileLabel.textContent = selectedFile?.name ?? "ファイルを選択";
        updateConvertButton();
    });

    elements.convertBtn.addEventListener("click", async () => {
        elements.convertBtn.disabled = true;
        await convertFile();
        updateConvertButton();
    });

    function setStatus(message, type = "info") {
        elements.status.textContent = message;
        elements.status.dataset.type = type;
    }

    function updateConvertButton() {
        elements.convertBtn.disabled = !wasmReady || !selectedFile;
    }

    async function loadWasm() {
        setStatus("WASM を読み込んでいます…");
        try {
            const result = await WebAssembly.instantiateStreaming(fetch("jww-dxf.wasm"), go.importObject).catch(async () => {
                const response = await fetch("jww-dxf.wasm");
                const bytes = await response.arrayBuffer();
                return WebAssembly.instantiate(bytes, go.importObject);
            });
            go.run(result.instance);
            wasmReady = true;
            setStatus("WASM がロードされました。JWW ファイルを選択してください。", "success");
            updateConvertButton();
        } catch (error) {
            console.error(error);
            setStatus(`WASM のロードに失敗しました: ${error.message}`, "error");
        }
    }

    async function convertFile() {
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

            const dxfDocResult = jwwToDxf(bytes);
            if (!dxfDocResult.ok) {
                throw new Error(dxfDocResult.error);
            }
            const doc = JSON.parse(dxfDocResult.data);
            updateMeta(doc);

            setStatus("DXF を生成しています…");
            const dxfStringResult = jwwToDxfString(bytes);
            if (!dxfStringResult.ok) {
                throw new Error(dxfStringResult.error);
            }

            const dxfString = dxfStringResult.data;
            elements.dxfOutput.value = dxfString;
            updateDownloadLink(dxfString);
            renderPreview(dxfString);
            setStatus("DXF を生成しました。プレビューとダウンロードが利用できます。", "success");
        } catch (error) {
            console.error(error);
            setStatus(`変換に失敗しました: ${error.message}`, "error");
            resetViewer("プレビューを表示できませんでした。");
            elements.downloadLink.classList.add("disabled");
            elements.downloadLink.setAttribute("aria-disabled", "true");
            elements.downloadLink.removeAttribute("href");
        }
    }

    function updateMeta(doc) {
        const entries = [
            { label: "レイヤー数", value: doc?.Layers?.length ?? 0 },
            { label: "エンティティ数", value: doc?.Entities?.length ?? 0 },
            { label: "ブロック数", value: doc?.Blocks?.length ?? 0 },
        ];

        elements.meta.innerHTML = entries
            .map(
                (entry) =>
                    `<div><dt>${entry.label}</dt><dd>${entry.value.toLocaleString()}</dd></div>`
            )
            .join("");
    }

    function updateDownloadLink(dxfString) {
        if (downloadUrl) {
            URL.revokeObjectURL(downloadUrl);
        }
        const blob = new Blob([dxfString], { type: "application/dxf" });
        downloadUrl = URL.createObjectURL(blob);

        const filename = selectedFile?.name?.replace(/\.jww$/i, "") || "output";
        elements.downloadLink.href = downloadUrl;
        elements.downloadLink.download = `${filename}.dxf`;
        elements.downloadLink.classList.remove("disabled");
        elements.downloadLink.setAttribute("aria-disabled", "false");
    }

    function renderPreview(dxfString) {
        resetViewer("プレビューを準備しています…");

        if (!window.DxfParser || !window.ThreeDxf || !window.THREE) {
            elements.viewerMessage.textContent =
                "プレビューライブラリを読み込めませんでした。ネットワーク接続を確認してください。";
            return;
        }

        try {
            const parser = new window.DxfParser();
            const parsed = parser.parseSync(dxfString);
            const width = elements.viewer.clientWidth || 640;
            const height = elements.viewer.clientHeight || 480;
            const viewer = new window.ThreeDxf.Viewer(parsed, elements.viewer, width, height);
            viewer.render();
            elements.viewerMessage.textContent = "右クリックでパン、マウスホイールでズームできます。";
        } catch (error) {
            console.error(error);
            resetViewer(`DXF プレビューの生成に失敗しました: ${error.message}`);
        }
    }

    function resetViewer(message) {
        elements.viewer.innerHTML = "";
        elements.viewer.appendChild(elements.viewerMessage);
        elements.viewerMessage.textContent = message;
    }
})();
