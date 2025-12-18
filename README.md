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

## Conversion Statistics

実行したjwwファイルは[Jw_cad](https://www.jwcad.net/)に同梱されているファイルを使用した。

### Test Data Matrix

| File            | Version | Line | Arc | Point | Text | Solid | Block | BlockDef | Error |
|---              |---      |---   |---  |---    |---   |---    |---    |---       |---    |
| `Test1.jww`     | 600     | 1642 | 4   | 4     | 36   | 0     | 0     | 0        |       |
| `Test2.jww`     | 600     | 38   | 4   | 6     | 23   | 0     | 0     | 0        |       |
| `Test3.jww`     | 600     | 104  | 6   | 33    | 56   | 0     | 0     | 0        |       |
| `Test4.jww`     | 600     | 72   | 1   | 18    | 21   | 0     | 0     | 0        |       |
| `Test5.jww`     | 600     | 46   | 0   | 0     | 43   | 0     | 0     | 0        |       |
| `Test6.jww`     | 600     | 1641 | 67  | 19    | 235  | 0     | 0     | 0        |       |
| `Test7.jww`     | 600     | 4083 | 5   | 26    | 93   | 0     | 0     | 0        |       |
| `サンプル.jww`      | 600     | 592  | 64  | 19    | 20   | 0     | 0     | 0        |       |
| `天空率表.jww`      | 600     | 6037 | 35  | 52    | 351  | 0     | 0     | 0        |       |
| `敷地図.jww`       | 600     | 9    | 0   | 0     | 0    | 0     | 0     | 0        |       |
| `日影図.jww`       | 600     | 1385 | 6   | 50    | 299  | 0     | 0     | 0        |       |
| `木造平面例.jww`     | 600     | 751  | 20  | 0     | 9    | 0     | 0     | 0        |       |
| `Ａマンション25d.jww` | 600     | 2150 | 0   | 12    | 70   | 0     | 0     | 0        |       |
| `Ａマンション平面例.jww` | 600     | 1394 | 215 | 10    | 56   | 79    | 0     | 0        |       |
| `Ａマンション立面例.jww` | 600     | 612  | 0   | 0     | 1    | 0     | 0     | 0        |       |

### DXF Conversion Results (Entity Count Comparison)

| File            | JWW Entities | DXF Entities | Diff | Status |
|---              |---           |---           |---   |---     |
| `Test1.jww`     | 1686         | 1686         | 0 ✅  | ✅      |
| `Test2.jww`     | 71           | 71           | 0 ✅  | ✅      |
| `Test3.jww`     | 199          | 188          | -11  | ✅      |
| `Test4.jww`     | 112          | 103          | -9   | ✅      |
| `Test5.jww`     | 89           | 89           | 0 ✅  | ✅      |
| `Test6.jww`     | 1962         | 1962         | 0 ✅  | ✅      |
| `Test7.jww`     | 4207         | 4207         | 0 ✅  | ✅      |
| `サンプル.jww`      | 695          | 695          | 0 ✅  | ✅      |
| `天空率表.jww`      | 6475         | 6457         | -18  | ✅      |
| `敷地図.jww`       | 9            | 9            | 0 ✅  | ✅      |
| `日影図.jww`       | 1740         | 1720         | -20  | ✅      |
| `木造平面例.jww`     | 780          | 780          | 0 ✅  | ✅      |
| `Ａマンション25d.jww` | 2232         | 2232         | 0 ✅  | ✅      |
| `Ａマンション平面例.jww` | 1754         | 1754         | 0 ✅  | ✅      |
| `Ａマンション立面例.jww` | 613          | 613          | 0 ✅  | ✅      |

### ezdxf Audit Results

| File            | Errors | Fixes | Status |
|---              |---     |---    |---     |
| `Test1.jww`     | 0      | 0     | ✅      |
| `Test2.jww`     | 0      | 0     | ✅      |
| `Test3.jww`     | 0      | 0     | ✅      |
| `Test4.jww`     | 0      | 0     | ✅      |
| `Test5.jww`     | 0      | 0     | ✅      |
| `Test6.jww`     | 0      | 0     | ✅      |
| `Test7.jww`     | 0      | 0     | ✅      |
| `サンプル.jww`      | 0      | 0     | ✅      |
| `天空率表.jww`      | 0      | 0     | ✅      |
| `敷地図.jww`       | 0      | 0     | ✅      |
| `日影図.jww`       | 0      | 0     | ✅      |
| `木造平面例.jww`     | 0      | 0     | ✅      |
| `Ａマンション25d.jww` | 0      | 0     | ✅      |
| `Ａマンション平面例.jww` | 0      | 0     | ✅      |
| `Ａマンション立面例.jww` | 0      | 0     | ✅      |

### ezdxf Info Results (DXF File Statistics)

| File            | Entities | Layers | Blocks | Status |
|---              |---       |---     |---     |---     |
| `Test1.jww`     | 1686     | 258    | 2      | ✅      |
| `Test2.jww`     | 71       | 258    | 2      | ✅      |
| `Test3.jww`     | 188      | 258    | 2      | ✅      |
| `Test4.jww`     | 103      | 258    | 2      | ✅      |
| `Test5.jww`     | 89       | 258    | 2      | ✅      |
| `Test6.jww`     | 1962     | 258    | 2      | ✅      |
| `Test7.jww`     | 4207     | 258    | 2      | ✅      |
| `サンプル.jww`      | 695      | 258    | 2      | ✅      |
| `天空率表.jww`      | 6457     | 258    | 2      | ✅      |
| `敷地図.jww`       | 9        | 258    | 2      | ✅      |
| `日影図.jww`       | 1720     | 258    | 2      | ✅      |
| `木造平面例.jww`     | 780      | 258    | 2      | ✅      |
| `Ａマンション25d.jww` | 2232     | 258    | 2      | ✅      |
| `Ａマンション平面例.jww` | 1754     | 258    | 2      | ✅      |
| `Ａマンション立面例.jww` | 613      | 258    | 2      | ✅      |

### Summary

- Total files: 15
- Successfully parsed: 15
- Parse errors: 0
- Successfully converted to DXF: 15
- ezdxf audit passed (0 errors): 15
- ezdxf total fixes applied: 0


## 開発

### JWW

JWWファイルの解析結果を表示
```bash
go run ./cmd/jww-stats/ examples/jww
```

### ビルド

```bash
make build       # ネイティブバイナリのビルド
make build-wasm  # WebAssembly のビルド
make test        # テストの実行
```

## 検証

* [ODA File Converter](https://www.opendesign.com/guestfiles/oda_file_Converter)
* [ezdxf](https://github.com/mozman/ezdxf)

## ライセンス

このプロジェクトは [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html) の下で提供されます。詳細は [LICENSE](./LICENSE) を参照してください。

## 既知の課題

### ODA FileConverter 互換性

出力したDXFファイルはezdxf auditでは正常に読み込めますが、ODA FileConverterでDWGに変換する際、以下のエラーが発生します：

- `Record name is empty - Ignored` (レイヤーテーブル)
- `Syntax error or premature end of file`
- `Null object Id`

原因調査中です。DXFの基本的な構造（HEADER、TABLES、ENTITIES）は正しいですが、ODAがより厳格なDXF構造を期待している可能性があります。
