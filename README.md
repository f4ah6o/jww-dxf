# jww-dxf

Jw_cad (JWW) ファイルを解析し、DXF 形式への変換や情報抽出を行うための Go 言語ライブラリおよびツールです。

## 特徴

- **JWW パーサー**: JWW ファイルのバイナリ構造を解析し、Go の構造体に変換します。
- **DXF エクスポート**: 解析したデータを DXF 形式で出力可能です。
- **WebAssembly 対応**: ブラウザ上での動作を想定した WASM ビルドをサポートしています。
- **コマンドラインツール**: ファイルの情報を表示したり、DXF に変換したりする CLI ツールが含まれています。

## インストール

Go がインストールされている環境で以下を実行してください。

```bash
go get github.com/f4ah6o/jww-dxf
```

## 使用方法

### CLI ツール

バイナリのビルド:
```bash
make build
```

JWW ファイルの情報を表示:
```bash
./bin/jww-dxf input.jww
```

DXF 形式で出力:
```bash
./bin/jww-dxf -dxf -o output.dxf input.jww
```

### ライブラリとしての利用

```go
import (
    "github.com/f4ah6o/jww-dxf/jww"
    "os"
)

func main() {
    f, _ := os.Open("example.jww")
    defer f.Close()

    doc, err := jww.Parse(f)
    if err != nil {
        panic(err)
    }
    // doc を使用してデータを処理
}
```

## 開発

### ビルド

```bash
make build       # ネイティブバイナリのビルド
make build-wasm  # WebAssembly のビルド
make test        # テストの実行
```

## 検証

* [ODA File Converter](https://www.opendesign.com/guestfiles/oda_file_Converter)
* [ezdxf](https://github.com/mozman/ezdxf)
