/**
 * 是否应把 SFTP 进度事件写入传输面板。
 * 历史上仅处理 upload，导致下载有事件但不更新进度条。
 */
export function shouldApplyTransferProgress(kind) {
  return kind === 'upload' || kind === 'download' || kind == null || kind === '';
}
