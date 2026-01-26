import { useState } from 'react';
import { Search, Server, Folder, Plus, Edit2, Trash2, ChevronDown, ChevronRight, Database, Container } from 'lucide-react';

export interface ServerAsset {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  group: string;
  tags?: string[];
  status?: 'online' | 'offline';
  type?: 'ssh' | 'database' | 'docker';
  dbType?: string;
  database?: string;
}

interface AssetListProps {
  onConnect: (asset: ServerAsset) => void;
  onAddClick: () => void;
  assets: ServerAsset[];
}

export function AssetList({ onConnect, onAddClick, assets }: AssetListProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set(['生产环境', '开发环境']));

  const filteredAssets = assets.filter(asset => {
    const searchLower = searchTerm.toLowerCase();
    return (
      asset.name.toLowerCase().includes(searchLower) ||
      asset.host.toLowerCase().includes(searchLower) ||
      asset.group.toLowerCase().includes(searchLower) ||
      asset.username.toLowerCase().includes(searchLower)
    );
  });

  const groupedAssets = filteredAssets.reduce((acc, asset) => {
    if (!acc[asset.group]) {
      acc[asset.group] = [];
    }
    acc[asset.group].push(asset);
    return acc;
  }, {} as Record<string, ServerAsset[]>);

  const toggleGroup = (group: string) => {
    const newExpanded = new Set(expandedGroups);
    if (newExpanded.has(group)) {
      newExpanded.delete(group);
    } else {
      newExpanded.add(group);
    }
    setExpandedGroups(newExpanded);
  };

  return (
    <div className="h-full flex flex-col bg-white border-r border-gray-200">
      {/* 头部 */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-900">服务器资产</h2>
          <button 
            onClick={onAddClick}
            className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <Plus className="w-4 h-4 text-purple-600" />
          </button>
        </div>
        
        {/* 搜索框 */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="搜索服务器..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
          />
        </div>
      </div>

      {/* 资产列表 */}
      <div className="flex-1 overflow-y-auto">
        {Object.entries(groupedAssets).map(([group, groupAssets]) => (
          <div key={group} className="mb-1">
            {/* 分组头部 */}
            <div
              onClick={() => toggleGroup(group)}
              className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer transition-colors"
            >
              {expandedGroups.has(group) ? (
                <ChevronDown className="w-4 h-4 text-gray-500" />
              ) : (
                <ChevronRight className="w-4 h-4 text-gray-500" />
              )}
              <Folder className="w-4 h-4 text-amber-500" />
              <span className="text-sm font-medium text-gray-700">{group}</span>
              <span className="ml-auto text-xs text-gray-400">({groupAssets.length})</span>
            </div>

            {/* 分组内的服务器 */}
            {expandedGroups.has(group) && (
              <div className="ml-4">
                {groupAssets.map((asset) => {
                  const getAssetIcon = () => {
                    switch (asset.type) {
                      case 'database':
                        return <Database className="w-4 h-4 text-blue-600 flex-shrink-0" />;
                      case 'docker':
                        return <Container className="w-4 h-4 text-cyan-600 flex-shrink-0" />;
                      default:
                        return <Server className="w-4 h-4 text-purple-600 flex-shrink-0" />;
                    }
                  };

                  return (
                    <div
                      key={asset.id}
                      onClick={() => onConnect(asset)}
                      className="group flex items-center gap-2 px-3 py-2.5 hover:bg-purple-50 rounded-lg mx-2 cursor-pointer transition-all"
                    >
                      {getAssetIcon()}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-gray-900 truncate">{asset.name}</span>
                          <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
                            asset.status === 'online' ? 'bg-green-500' : 'bg-gray-300'
                          }`} />
                        </div>
                        <div className="text-xs text-gray-500 truncate">
                          {asset.username}@{asset.host}:{asset.port}
                          {asset.type === 'database' && asset.dbType && ` • ${asset.dbType.toUpperCase()}`}
                        </div>
                      </div>
                      <div className="opacity-0 group-hover:opacity-100 flex gap-1 transition-opacity">
                        <button className="p-1 hover:bg-purple-100 rounded">
                          <Edit2 className="w-3 h-3 text-gray-600" />
                        </button>
                        <button className="p-1 hover:bg-purple-100 rounded">
                          <Trash2 className="w-3 h-3 text-gray-600" />
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}