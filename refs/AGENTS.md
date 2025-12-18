# JWW Data Format Analysis

JWWファイル形式の解析知識をまとめたドキュメント。

## File Structure Overview

JWWファイルはバイナリ形式で、MFC (Microsoft Foundation Classes) の CArchive シリアライズを使用しています。

```
+-------------------+
| Signature (8B)    |  "JwwData."
+-------------------+
| Version (DWORD)   |  例: 600 = Ver 6.00, 700 = Ver 7.02
+-------------------+
| File Memo         |  CString (Shift-JIS)
+-------------------+
| Paper Size        |  DWORD (0-4: A0-A4, 8-11: 2A-5A, etc.)
+-------------------+
| Layer Groups      |  16グループ × 16レイヤ
+-------------------+
| Settings...       |  多数の設定値
+-------------------+
| Entity List       |  図形データ
+-------------------+
| Block Definitions |  ブロック定義
+-------------------+
| Embedded Images   |  Ver.7.00+ (zlib圧縮)
+-------------------+
```

## Data Types

| Type | Size | Description |
|------|------|-------------|
| DWORD | 4 bytes | uint32, little-endian |
| WORD | 2 bytes | uint16, little-endian |
| BYTE | 1 byte | uint8 |
| double | 8 bytes | float64, little-endian |
| CString | variable | Length-prefixed string (Shift-JIS) |

### CString Format

```
if length < 0xFF:
    [1 byte length] [string bytes]
else if length < 0xFFFF:
    [0xFF] [2 byte length] [string bytes]
else:
    [0xFF] [0xFF 0xFF] [4 byte length] [string bytes]
```

## Entity List Serialization

MFC CTypedPtrList<CObList, CData*> 形式：

```
[WORD: entity count]
[Entity 0]
[Entity 1]
...
[Entity N-1]
```

### Entity Format

各エンティティは以下の形式：

```
[WORD: class identifier]
    if 0xFFFF (new class):
        [WORD: schema version] (例: 0x0258 = 600)
        [WORD: class name length]
        [class name bytes] (例: "CDataSen")
    if 0x8000:
        null object
    else:
        existing class reference (mask with 0x7FFF)
[entity data...]
```

### Entity Base (CData)

全エンティティの基底クラス：

```go
type EntityBase struct {
    Group      uint32 // DWORD: 曲線属性番号
    PenStyle   byte   // BYTE: 線種番号
    PenColor   uint16 // WORD: 線色番号
    PenWidth   uint16 // WORD: 線幅 (Ver.3.51+)
    Layer      uint16 // WORD: レイヤ番号
    LayerGroup uint16 // WORD: レイヤグループ番号
    Flag       uint16 // WORD: 属性フラグ
}
```

## Entity Classes

### CDataSen (線)

```
[EntityBase]
[double: start.x]
[double: start.y]
[double: end.x]
[double: end.y]
```

### CDataEnko (円弧)

```
[EntityBase]
[double: center.x]
[double: center.y]
[double: radius]
[double: start_angle]    // radians
[double: arc_angle]      // radians
[double: tilt_angle]     // radians
[double: flatness]       // 扁平率 (1.0 = 真円)
[DWORD: is_full_circle]  // 0/1
```

### CDataTen (点)

```
[EntityBase]
[double: x]
[double: y]
[DWORD: is_temporary]
if PenStyle == 100:
    [DWORD: code]        // 矢印・ポイントマーカー
    [double: angle]
    [double: scale]
```

### CDataMoji (文字)

```
[EntityBase]
[double: start.x]
[double: start.y]
[double: end.x]
[double: end.y]
[DWORD: text_type]       // +10000: italic, +20000: bold
[double: size_x]
[double: size_y]
[double: spacing]
[double: angle]          // degrees
[CString: font_name]
[CString: content]
```

### CDataSolid (ソリッド)

```
[EntityBase]
[double: point1.x] [double: point1.y]
[double: point4.x] [double: point4.y]  // Note: 4 before 2,3
[double: point2.x] [double: point2.y]
[double: point3.x] [double: point3.y]
if PenColor == 10:
    [DWORD: rgb_color]
```

### CDataBlock (ブロック参照)

```
[EntityBase]
[double: ref.x]
[double: ref.y]
[double: scale_x]
[double: scale_y]
[double: rotation]       // radians
[DWORD: block_def_number]
```

### CDataSunpou (寸法)

```
[EntityBase]
[CDataSen: line]         // 寸法線
[CDataMoji: text]        // 寸法値
if version >= 420:       // SXF mode
    [WORD: sxf_mode]
    [CDataSen: helper_line_1]
    [CDataSen: helper_line_2]
    [CDataTen: point_1]
    [CDataTen: point_2]
    [CDataTen: base_point_1]
    [CDataTen: base_point_2]
```

## Attribute Flags (m_sFlg)

| Bit | 線 | 円 | 点 | 文字 | ブロック | 寸法 | ソリッド |
|-----|----|----|----|----|----------|------|---------|
| 0x0010 | | 図形 | 図形 | 寸法値 | | | |
| 0x0020 | ハッチ | ハッチ | ハッチ | 縦字 | | | ハッチ |
| 0x0040 | | 寸法 | 寸法 | 真北 | | | |
| 0x0080 | | 建具 | 建具 | 日影 | | | |
| 0x0100 | | | | 半径寸法値 | | | |
| 0x0200 | | | | 直径寸法値 | | | |
| 0x0400 | | | | 角度寸法値 | | | |
| 0x0800 | 図形 | | | 図形属性選択 | 図形 | | |
| 0x1000 | 建具 | | | 累計寸法値 | 建具 | | |
| 0x2000 | 寸法 | | | 建具 | 寸法 | 寸法 | |
| 0x4000 | | | | 寸法 | | | |
| 0x8000 | 包絡処理対象外建具 | | | 2.5D | | | |

## Version-Specific Notes

- **Ver.3.00+**: マークジャンプ8セット、天空図設定
- **Ver.3.51+**: EntityBaseに線幅(WORD)追加
- **Ver.4.20+**: SXF拡張線色/線種定義、寸法SXFモード
- **Ver.6.00+**: 線描画最大幅の計算変更
- **Ver.7.00+**: 同梱画像ファイル対応

## Implementation Notes

1. **バイトオーダー**: 全てリトルエンディアン
2. **文字コード**: Shift-JIS → UTF-8への変換が必要
3. **エンティティカウント**: WORD (2バイト) で格納
4. **クラスID**: 出現順に1から割り当て、参照時は最上位ビットなし
5. **座標系**: mm単位、Y軸上向き

## Known Issues

- EntityBaseのサイズはバージョンにより変動 (Ver.3.51前後で2バイト差)
- MFCクラス参照のインデックスは実装依存の可能性あり
- 一部の設定フィールドのサイズが不明確
