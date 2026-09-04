# Bug：常用命令记录成 Tab 半截按键

## 背景

用户用 Tab 做 shell 补全（如 `cd lnpe` + Tab → `cd lnpems-idc-hiddendanger/`）后回车，历史里却记成 `cd lnpe -id` 这类半成品。根因是本地按键缓冲把 `\t` 当字符写入，而补全展开只出现在远端回显 / xterm 画面上。

## 范围

- `commandLineBuffer`：忽略 Tab，提供 `replaceLine`
- 新增 `terminalCommandLine.js`：从 xterm 光标行剥离 prompt 得到真实命令
- `Terminal.svelte`：回车优先用可见行 `RecordCommand`；Tab 后短暂延迟同步缓冲

## 修改文件

- `frontend/src/lib/commandLineBuffer.js`
- `frontend/src/lib/terminalCommandLine.js`
- `frontend/src/components/Terminal.svelte`
- `frontend/test/commandLineBuffer.test.js`
- `frontend/test/terminalCommandLine.test.js`

## 验证

```bash
cd frontend && node --test test/commandLineBuffer.test.js test/terminalCommandLine.test.js test/commandSuggest.test.js
```

## 剩余风险

- Prompt 形态极端非常规时，`extractShellCommand` 可能剥不准；多行粘贴仍走按键缓冲
- Tab 后建议浮层依赖约 80ms 回显同步，高延迟链路可能短暂仍用旧前缀查询
