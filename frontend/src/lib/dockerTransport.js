/**
 * Build metadata map for a docker connection asset (backend uses snake_case keys).
 */
export function buildDockerMetadata({
  transport = 'local',
  sshConnectionId,
  socketPath,
  remoteSocket,
  cliFallback,
  tls
} = {}) {
  const metadata = {
    domain: 'docker',
    transport
  };

  if (sshConnectionId) {
    metadata.ssh_connection_id = sshConnectionId;
  }
  if (socketPath) {
    metadata.socket_path = socketPath;
  }
  if (remoteSocket) {
    metadata.remote_socket = remoteSocket;
  }
  if (cliFallback === false) {
    metadata.cli_fallback = 'false';
  } else if (cliFallback === true) {
    metadata.cli_fallback = 'true';
  }
  if (tls === true) {
    metadata.tls = 'true';
  } else if (tls === false) {
    metadata.tls = 'false';
  }

  return metadata;
}

/** Default display name when adding Docker from an SSH asset. */
export function defaultDockerName(sshAsset) {
  return `${sshAsset?.name ?? ''}-Docker`;
}
