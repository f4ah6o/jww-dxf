# リポジトリ運用情報

## バージョニング
`v<yyyy>.<MM>.<increment>` の形式を採用しています。

## リリース（Publish）
`tag push` トリガーで自動的に publish されます。

---

# DXFファイル構造リファレンス

このドキュメントは、AutoCAD 2024 Developer and ObjectARX HelpからのDXFファイル構造に関する情報をまとめたものです。

**出典**: [AutoCAD 2024 Developer and ObjectARX Help - DXF Reference](https://help.autodesk.com/view/OARX/2024/ENU/)

## 1. DXFファイルの概要

DXF (Drawing Interchange Format) ファイルは、グループコードと関連する値のペアで構成されています。グループコードは、後続の値の種類を示します。

### 基本構造

* グループコードと値がペアで記述される
* 各グループコードと値は独立した行に配置される
* セクション単位で構成される
* 各セクションは `0 SECTION` で始まり、`0 ENDSEC` で終わる

### 文字列長の制限

* 最大DXFファイル文字列長: 256文字
* この制限を超える文字列は、SAVE、SAVEAS、WBLOCKの際に切り詰められる
* OPENとINSERTは、256文字を超える文字列を含むDXFファイルで失敗する

## 2. DXFファイルの全体構成

DXFファイルは以下のセクションで構成されます:

### 2.1 HEADER セクション

* 図面に関する一般情報を含む
* データベースバージョン番号と多数のシステム変数で構成
* 各パラメータは変数名と関連値を含む

**構造例**:
```
0
SECTION
2
HEADER
9
$<variable>
<group code>
<value>
0
ENDSEC
```

### 2.2 CLASSES セクション

* アプリケーション定義クラスの情報を保持
* クラスのインスタンスはBLOCKS、ENTITIES、OBJECTSセクションに表示される
* クラス定義はクラス階層内で永続的に固定される

**構造例**:
```
0
SECTION
2
CLASSES
0
CLASS
1
<class dxf record>
2
<class name>
3
<app name>
90
<flag>
280
<flag>
281
<flag>
0
ENDSEC
```

### 2.3 TABLES セクション

以下のシンボルテーブルの定義を含む:

* APPID (アプリケーション識別テーブル)
* BLOCK_RECORD (ブロック参照テーブル)
* DIMSTYLE (寸法スタイルテーブル)
* LAYER (レイヤーテーブル)
* LTYPE (線種テーブル)
* STYLE (テキストスタイルテーブル)
* UCS (ユーザー座標系テーブル)
* VIEW (ビューテーブル)
* VPORT (ビューポート構成テーブル)

**構造例**:
```
0
SECTION
2
TABLES
0
TABLE
2
<table type>
5
<handle>
100
AcDbSymbolTable
70
<max. entries>
0
<table type>
5
<handle>
100
AcDbSymbolTableRecord
.
. <data>
.
0
ENDTAB
0
ENDSEC
```

### 2.4 BLOCKS セクション

* ブロック定義と、各ブロック参照を構成する図形エンティティを含む

**構造例**:
```
0
SECTION
2
BLOCKS
0
BLOCK
5
<handle>
100
AcDbEntity
8
<layer>
100
AcDbBlockBegin
2
<block name>
70
<flag>
10
<X value>
20
<Y value>
30
<Z value>
3
<block name>
1
<xref path>
0
<entity type>
.
. <data>
.
0
ENDBLK
5
<handle>
100
AcDbBlockEnd
0
ENDSEC
```

### 2.5 ENTITIES セクション

* 図面内のグラフィカルオブジェクト（エンティティ）を含む
* ブロック参照（挿入エンティティ）を含む

**構造例**:
```
0
SECTION
2
ENTITIES
0
<entity type>
5
<handle>
330
<pointer to owner>
100
AcDbEntity
8
<layer>
100
AcDb<classname>
.
. <data>
.
0
ENDSEC
```

**注**: SAVEまたはSAVEASコマンドのSelect Objectsオプションを使用すると、結果のDXFファイルのENTITIESセクションには選択したエンティティのみが含まれます。

### 2.6 OBJECTS セクション

* 図面内の非グラフィカルオブジェクトを含む
* エンティティ、シンボルテーブルレコード、シンボルテーブル以外のすべてのオブジェクトが格納される
* 例: mlineスタイルとグループを含む辞書

**構造例**:
```
0
SECTION
2
OBJECTS
0
DICTIONARY
5
<handle>
100
AcDbDictionary
3
<dictionary name>
350
<handle of child>
0
<object type>
.
. <data>
.
0
ENDSEC
```

### 2.7 THUMBNAILIMAGE セクション

* 図面のプレビュー画像データを含む
* このセクションはオプション

## 3. グループコード

グループコードと関連する値は、オブジェクトまたはエンティティの特定の側面を定義します。

### グループコードの特性

* グループコードの直後の行が関連する値
* 値は文字列、整数、または浮動小数点値（点のX座標など）
* ファイルセパレータとして特別なグループコードが使用される

### エンティティとオブジェクトの導入

* エンティティ、オブジェクト、クラス、テーブル、テーブルエントリ、ファイルセパレータは、0グループコードで導入される
* 0グループコードの後に、グループを説明する名前が続く

## 4. DXFインターフェースプログラムの作成

AutoCADベースのプログラムとDXFファイルを介して通信するプログラムの作成は、見た目よりも簡単です。DXF形式により、必要な情報を読み取りながら、不要な情報を簡単に無視できます。

### 主要な考慮事項

* **読み取り**: 必要な情報のみを選択的に読み取る
* **書き込み**: AutoCADが認識できる適切な形式でデータを出力する
* **情報の無視**: DXF形式により、不要な情報を簡単にスキップできる

## 5. 参考資料

各セクションの詳細については、以下のAutodesk公式ドキュメントを参照してください:

* [About the General DXF File Structure](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-D939EA11-0CEC-4636-91A8-756640A031D3)
* [About Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-89CB823D-614D-4D1E-8204-568EC72DF869)
* [Header Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-2A01D125-C1C9-4B20-B916-0F5598C8F19E)
* [Class Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-42E19B4F-61E1-4795-93E7-C8769CE2D7C0)
* [Symbol Table Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-A66D0ACA-3F43-4B2E-A0C2-2B490C1E5268)
* [Blocks Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-A3E4F1D3-79C9-489C-B7EC-3924DA7F25C9)
* [Entity Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-995ABB55-571A-4D0F-882E-8A74A738643E)
* [Object Group Codes in DXF Files](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-1038FDE4-745D-469D-972E-1F977D674882)
* [About Writing a DXF Interface Program](https://help.autodesk.com/view/OARX/2024/ENU/?guid=GUID-3CBD1CD7-D928-4630-A46E-F8408F761FCF)

---

**ライセンス**: このドキュメントは、[Creative Commons Attribution-NonCommercial-ShareAlike 3.0 Unported License](https://creativecommons.org/licenses/by-nc-sa/3.0/)のもとでライセンスされている情報に基づいています。

**著作権**: © 2025 Autodesk Inc. All rights reserved
