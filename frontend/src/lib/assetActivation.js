export function shouldConnectAsset(event) {
  return event?.type === 'dblclick' || (event?.type === 'keydown' && event.key === 'Enter');
}
