import { uniqueCopyName } from './fileManagerContextMenu.js';

export function normalizeUploadItems(raw) {
  return (Array.isArray(raw) ? raw : [])
    .map((item) => ({
      localPath: String(item?.localPath || item?.LocalPath || '').trim(),
      relPath: String(item?.relPath || item?.RelPath || '').trim().replace(/\\/g, '/'),
      isDir: Boolean(item?.isDir ?? item?.IsDir),
    }))
    .filter((item) => item.localPath && item.relPath);
}

export function classifyConflict(item, remote) {
  if (!item || !remote) return null;
  if (item.isDir && remote.isDir) return null;
  if (!item.isDir && !remote.isDir) return 'file';
  return 'type';
}

export function indexRemoteEntries(relDir, entries, remoteIndex = new Map()) {
  for (const entry of entries || []) {
    if (entry?.is_parent) continue;
    const name = entry.name || entry.Name;
    if (!name || name === '.' || name === '..') continue;
    const rel = relDir ? `${relDir}/${name}` : name;
    remoteIndex.set(rel, { isDir: Boolean(entry.is_dir ?? entry.IsDir) });
  }
  return remoteIndex;
}

export async function buildRemoteIndex(items, listDirectory) {
  const remoteIndex = new Map();
  const listed = new Set();
  const queue = [''];

  while (queue.length) {
    const relDir = queue.shift();
    if (listed.has(relDir)) continue;
    listed.add(relDir);

    let entries = [];
    try {
      entries = await listDirectory(relDir);
    } catch {
      entries = [];
    }
    indexRemoteEntries(relDir, entries, remoteIndex);

    for (const item of items || []) {
      if (!item.isDir) continue;
      if (!remoteIndex.get(item.relPath)?.isDir) continue;
      if (!listed.has(item.relPath) && !queue.includes(item.relPath)) {
        queue.push(item.relPath);
      }
    }
  }

  return remoteIndex;
}

export function conflictsForItems(items, remoteIndex) {
  return (items || []).flatMap((item) => {
    const kind = classifyConflict(item, remoteIndex.get(item.relPath));
    return kind ? [{ ...item, kind }] : [];
  });
}

export function remapPrefix(relPath, fromPrefix, toPrefix) {
  if (relPath === fromPrefix) return toPrefix;
  if (fromPrefix && relPath.startsWith(`${fromPrefix}/`)) {
    return `${toPrefix}${relPath.slice(fromPrefix.length)}`;
  }
  return relPath;
}

export function parentRel(relPath) {
  const parts = String(relPath || '').split('/').filter(Boolean);
  parts.pop();
  return parts.join('/');
}

export function leafName(relPath) {
  const parts = String(relPath || '').split('/').filter(Boolean);
  return parts[parts.length - 1] || String(relPath || '');
}

export function namesInParent(remoteIndex, parentRelPath) {
  const names = new Set();
  const prefix = parentRelPath ? `${parentRelPath}/` : '';
  const keys = remoteIndex?.keys ? remoteIndex.keys() : [];
  for (const rel of keys) {
    if (parentRelPath) {
      if (!rel.startsWith(prefix)) continue;
      const rest = rel.slice(prefix.length);
      if (rest && !rest.includes('/')) names.add(rest);
      continue;
    }
    if (rel && !rel.includes('/')) names.add(rel);
  }
  return names;
}

function occupiedNames(takenRelPaths, plan, parentRelPath) {
  const names = namesInParent(takenRelPaths, parentRelPath);
  for (const item of plan) {
    if (parentRel(item.relPath) === parentRelPath) names.add(leafName(item.relPath));
  }
  return names;
}

function withoutPrefix(items, prefix) {
  return items.filter((item) => item.relPath !== prefix && !item.relPath.startsWith(`${prefix}/`));
}

function renameTree(items, fromPrefix, toPrefix) {
  return items.map((item) => ({
    ...item,
    relPath: remapPrefix(item.relPath, fromPrefix, toPrefix),
  }));
}

function isUnderPrefix(relPath, prefix) {
  return relPath === prefix || relPath.startsWith(`${prefix}/`);
}

export async function resolveUploadConflicts(items, remoteIndex, prompts = {}) {
  const promptFolder = prompts.promptFolder || (async () => 'cancel');
  const promptFile = prompts.promptFile || (async () => 'cancel');
  const index = new Map(remoteIndex);
  const taken = new Set(index.keys());
  let plan = (items || []).map((item) => ({ ...item }));
  const deleteRemote = [];

  const topFolders = [...new Set(
    plan.filter((item) => item.isDir && !String(item.relPath).includes('/')).map((item) => item.relPath),
  )];

  for (const prefix of topFolders) {
    if (!index.get(prefix)?.isDir) continue;
    const group = plan.filter((item) => isUnderPrefix(item.relPath, prefix));
    const nested = conflictsForItems(group, index);
    if (nested.length === 0) continue;

    const decision = await promptFolder({
      name: prefix,
      relPath: prefix,
      conflictCount: nested.length,
    });

    if (decision === 'cancel') {
      plan = withoutPrefix(plan, prefix);
      continue;
    }

    if (decision === 'rename') {
      const newName = uniqueCopyName(occupiedNames(taken, plan, ''), prefix);
      plan = renameTree(plan, prefix, newName);
      taken.add(newName);
      continue;
    }

    if (decision === 'overwriteAll') {
      for (const conflict of nested) {
        if (conflict.kind === 'type') deleteRemote.push(conflict.relPath);
        index.delete(conflict.relPath);
      }
    }
  }

  const skipPrefixes = [];
  let applyAll = null;

  const isSkipped = (relPath) => skipPrefixes.some((prefix) => isUnderPrefix(relPath, prefix));

  while (true) {
    const leftover = conflictsForItems(plan.filter((item) => !isSkipped(item.relPath)), index);
    if (leftover.length === 0) break;

    const current = leftover[0];
    let action = applyAll;
    if (!action) {
      action = await promptFile({
        name: leafName(current.relPath),
        relPath: current.relPath,
        kind: current.kind,
        remaining: leftover.length,
        isDir: current.isDir,
      });
      if (action === 'overwriteAll') {
        applyAll = 'overwrite';
        action = 'overwrite';
      } else if (action === 'renameAll') {
        applyAll = 'rename';
        action = 'rename';
      }
    }

    if (action === 'cancel') {
      skipPrefixes.push(current.relPath);
      continue;
    }

    if (action === 'rename') {
      const parent = parentRel(current.relPath);
      const newLeaf = uniqueCopyName(occupiedNames(taken, plan, parent), leafName(current.relPath));
      const newRel = parent ? `${parent}/${newLeaf}` : newLeaf;
      plan = current.isDir
        ? renameTree(plan, current.relPath, newRel)
        : plan.map((item) => (
          item.localPath === current.localPath && item.relPath === current.relPath
            ? { ...item, relPath: newRel }
            : item
        ));
      index.delete(current.relPath);
      taken.add(newRel);
      continue;
    }

    if (current.kind === 'type') deleteRemote.push(current.relPath);
    index.delete(current.relPath);
  }

  return {
    items: plan.filter((item) => !isSkipped(item.relPath)),
    deleteRemote,
  };
}
