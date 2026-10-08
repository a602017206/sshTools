# 缺陷：建表/表结构字段类型不全

## 背景

表设计器的字段类型下拉框写死为 11 个通用类型（`BIGINT`、`INT`、`NUMBER`、`VARCHAR`、`VARCHAR2`、`TEXT`、`CLOB`、`DECIMAL`、`TIMESTAMP`、`DATE`、`BOOLEAN`），所有数据库共用：

- MySQL 缺少 `LONGTEXT`、`DATETIME`、`TINYINT`、`JSON` 等常用类型，同时混入了 `VARCHAR2`、`CLOB` 等 Oracle 类型。
- 打开已有表时，字段类型若不在列表中（如 `LONGTEXT`），下拉框显示为空白。

## 范围

- 新增 `columnTypeOptions(databaseType, currentType)`，按方言返回类型列表：
  - MySQL：整数、浮点、`BIT`、字符串、`TINYTEXT`～`LONGTEXT`、二进制、`TINYBLOB`～`LONGBLOB`、`DATE`/`TIME`/`DATETIME`/`TIMESTAMP`/`YEAR`、`JSON`、`BOOLEAN`。
  - PostgreSQL/人大金仓/openGauss：`SERIAL`、`BIGSERIAL`、`DOUBLE PRECISION`、`BYTEA`、`TIMESTAMPTZ`、`INTERVAL`、`JSONB`、`UUID` 等。
  - Oracle：`NUMBER`、`VARCHAR2`、`NVARCHAR2`、`CLOB`、`NCLOB`、`BLOB`、`RAW`、`BINARY_FLOAT` 等。
  - 其他数据库沿用原通用列表。
- 字段现有类型不在列表中时置顶保留，避免空白。
- 新增 `typeAcceptsLength`，建表和改表 SQL 共用；补充 `TINYINT`、`MEDIUMINT`、`BIT`、`BINARY`、`VARBINARY`、`NCHAR`、`RAW` 的长度拼接。
- 切换到不需要长度的类型（如 `LONGTEXT`、`DATETIME`）时自动清空长度，避免生成 `DATETIME(255)` 之类的非法语句。

## 修改文件

- `frontend/src/lib/columnTypeOptions.js`
- `frontend/src/lib/tableDefinitionSQL.js`
- `frontend/src/lib/tableAlterSQL.js`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/test/columnTypeOptions.test.js`
- `.github/workflows/release.yml`（发布说明）
- 本文档

## 验证

```bash
cd frontend && node --test test/*.test.js && npm run build
```

- 新增 4 个用例全部通过；全量 269 个中 267 个通过，失败的 2 个位于 `test/nativeDatabasePanelLayout.test.js`（Redis/Kafka 工作区），为既有失败，与本次无关。
- `vite build` 成功；构建脚本后续的 JDBC agent Gradle 打包步骤卡住（Gradle 守护进程问题，与本次前端改动无关），已手动终止。
- 未在桌面端对 MySQL 实际建表做手工验证。

## 剩余风险

- `ENUM`、`SET` 需要枚举值，下拉框无法填写，暂未提供；已有此类字段会按原类型显示，但改表时不会带上枚举值。
- MySQL `DATETIME`/`TIMESTAMP` 的小数秒精度（如 `DATETIME(3)`）暂不支持通过长度设置。
