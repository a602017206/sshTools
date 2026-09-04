export function resolvePreferredConnectionGroup(value) {
  if (typeof value !== 'string') return '';
  return value.trim();
}

export function applyPreferredGroupToFormData(formData, preferredGroup) {
  const group = resolvePreferredConnectionGroup(preferredGroup);
  if (!group) {
    return { ...formData };
  }
  return { ...formData, group };
}

export function splitGroupPath(group) {
  const raw = String(group ?? '');
  if (!raw.includes('/')) {
    return [raw];
  }
  const parts = raw.split('/').map((part) => part.trim()).filter(Boolean);
  return parts.length > 0 ? parts : [raw];
}

export function ancestorGroupPaths(path) {
  const parts = splitGroupPath(path).filter(Boolean);
  const ancestors = [];
  let current = '';
  for (const part of parts) {
    current = current ? `${current}/${part}` : part;
    ancestors.push(current);
  }
  return ancestors;
}

function compareGroupName(a, b) {
  return a.localeCompare(b, 'zh');
}

export function buildAssetGroupForest(groupedAssets) {
  const rootChildren = [];
  const index = new Map();

  function ensureNode(parts) {
    let parentList = rootChildren;
    let path = '';
    let node = null;
    for (const name of parts) {
      path = path ? `${path}/${name}` : name;
      node = index.get(path);
      if (!node) {
        node = { name, path, children: [], assets: [], assetCount: 0 };
        parentList.push(node);
        index.set(path, node);
      }
      parentList = node.children;
    }
    return node;
  }

  for (const [group, assets] of Object.entries(groupedAssets || {})) {
    const node = ensureNode(splitGroupPath(group));
    if (node) {
      node.assets.push(...(assets || []));
    }
  }

  function sortAndCount(nodes) {
    nodes.sort((left, right) => compareGroupName(left.name, right.name));
    let total = 0;
    for (const node of nodes) {
      node.assetCount = node.assets.length + sortAndCount(node.children);
      total += node.assetCount;
    }
    return total;
  }

  sortAndCount(rootChildren);
  return rootChildren;
}

export function flattenVisibleGroupTree(forest, expandedGroups) {
  const expanded = expandedGroups instanceof Set ? expandedGroups : new Set(expandedGroups || []);
  const rows = [];

  function walk(nodes, depth) {
    for (const node of nodes || []) {
      rows.push({ kind: 'group', node, depth });
      if (!expanded.has(node.path)) continue;
      walk(node.children, depth + 1);
      for (const asset of node.assets) {
        rows.push({ kind: 'asset', asset, depth: depth + 1, groupPath: node.path });
      }
    }
  }

  walk(forest, 0);
  return rows;
}

export function expandedGroupPathsForSearch(groupedAssets, searchTerm) {
  if (!String(searchTerm || '').trim()) {
    return [];
  }
  const paths = new Set();
  for (const group of Object.keys(groupedAssets || {})) {
    for (const path of ancestorGroupPaths(group)) {
      paths.add(path);
    }
  }
  return [...paths];
}
