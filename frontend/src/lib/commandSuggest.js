/** 选中建议后填入当前行（一期整行替换）。 */
export function pickSuggestFill(_currentLine, suggestion) {
  return suggestion;
}

/** 当前行非空且命令提示开关开启时才展示建议。 */
export function shouldOfferSuggest(line, enabled) {
  return Boolean(enabled) && line.trim() !== '';
}
