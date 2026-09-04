/** Enter 发送时忽略 IME 组字确认（中文输入法回车上屏 / 切英文）。 */
export function shouldSubmitComposerOnEnter(event) {
  if (!event || event.key !== 'Enter' || event.shiftKey) return false;
  if (event.isComposing || event.keyCode === 229) return false;
  return true;
}

export function isCopilotCancelError(error) {
  const text = String(error?.message || error || '').toLowerCase();
  return text.includes('cancel') || text.includes('context canceled') || text.includes('已取消') || text.includes('aborted');
}
