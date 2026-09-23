// vim / less / htop 把 LF 当作「下移一行、列不变」。
// convertEol 会在每次 LF 时把光标拉回第 0 列，状态栏就会散成一竖条碎片。
// 普通 shell 的换行由远端 PTY 的 ONLCR 转成 CRLF，这里不需要再转。
export function createXtermOptions({ fontSize, fontFamily, theme }) {
  return {
    cursorBlink: true,
    fontSize,
    fontFamily,
    theme,
    allowProposedApi: true,
    scrollback: 1000,
    scrollOnUserInput: true,
    convertEol: false,
    rightClickSelectsWord: false,
    macOptionClickForcesSelection: true
  };
}
