# 资产树连接状态灯与层级对齐

## 背景

1. SSH 服务器连接的状态灯始终为绿色，未连接或连接失败时也显示在线。
2. SSH 与数据库行的状态灯未对齐，SSH 行缺少展开位占位导致指示灯偏左。
3. 文件夹下的连接项缩进不足，层级感不明显。

## 范围

- 修正连接状态判断：仅依据实际会话/数据库连接状态，不再把持久化的 `status: online` 当作已连接
- 新建/加载连接默认状态改为 `idle`
- SSH 行增加与数据库展开按钮等宽的占位，统一状态灯列
- 增加分组内连接项缩进与左侧引导线

## 修改文件

- `frontend/src/lib/assetLinkState.js`（新增）
- `frontend/src/components/AssetList.svelte`
- `frontend/src/App.svelte`
- `frontend/src/styles/app.css`
- `frontend/test/assetLinkState.test.js`（新增）

## 验证

- [x] `node --test frontend/test/assetLinkState.test.js`
- [ ] 未连接 SSH 显示灰色状态灯
- [ ] 连接成功后显示绿色，连接中显示黄色
- [ ] SSH 与数据库状态灯垂直对齐
- [ ] 分组下连接项有明显缩进层次

## 剩余风险

- 连接失败后若未写入 `status: error`，状态灯会回到灰色而非红色；需后续在连接失败回调中补充 error 状态
