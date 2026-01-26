import { useState } from 'react';
import { AssetList, type ServerAsset } from '@/app/components/AssetList';
import { TerminalPanel } from '@/app/components/TerminalPanel';
import { FileManager } from '@/app/components/FileManager';
import { ServerMonitor } from '@/app/components/ServerMonitor';
import { Wrench } from 'lucide-react';
import { AddAssetDialog } from '@/app/components/AddAssetDialog';
import { DevToolsPanel } from '@/app/components/DevToolsPanel';

export default function App() {
  const [connections, setConnections] = useState<ServerAsset[]>([]);
  const [isDevToolsOpen, setIsDevToolsOpen] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  
  // 资产列表数据
  const [assets, setAssets] = useState<ServerAsset[]>([
    { id: '1', name: '生产服务器-01', host: '192.168.1.10', port: 22, username: 'root', group: '生产环境', status: 'online', type: 'ssh' },
    { id: '2', name: '生产服务器-02', host: '192.168.1.11', port: 22, username: 'root', group: '生产环境', status: 'online', type: 'ssh' },
    { id: '3', name: 'Web服务器', host: '192.168.1.20', port: 22, username: 'ubuntu', group: '生产环境', status: 'offline', type: 'ssh' },
    { id: '4', name: '开发服务器-01', host: '192.168.2.10', port: 22, username: 'dev', group: '开发环境', status: 'online', type: 'ssh' },
    { id: '5', name: '开发服务器-02', host: '192.168.2.11', port: 22, username: 'dev', group: '开发环境', status: 'online', type: 'ssh' },
    { id: '6', name: '测试服务器', host: '192.168.3.10', port: 22, username: 'test', group: '测试环境', status: 'online', type: 'ssh' },
    { id: '7', name: 'MySQL数据库', host: '192.168.1.30', port: 3306, username: 'root', group: '生产环境', status: 'online', type: 'database', dbType: 'mysql' },
    { id: '8', name: 'Docker主机', host: '192.168.1.40', port: 2375, username: 'docker', group: '开发环境', status: 'online', type: 'docker' },
  ]);

  const handleConnect = (asset: ServerAsset) => {
    // 检查是否已经连接
    if (!connections.find(c => c.id === asset.id)) {
      setConnections([...connections, asset]);
    }
  };

  const handleCloseConnection = (id: string) => {
    setConnections(connections.filter(c => c.id !== id));
  };

  const handleAddAsset = (newAsset: ServerAsset) => {
    setAssets([...assets, newAsset]);
  };

  return (
    <div className="w-full h-screen bg-gray-50 text-gray-900 flex flex-col">
      {/* 顶部标题栏 */}
      <div className="h-14 bg-white border-b border-gray-200 flex items-center px-6 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-gradient-to-br from-purple-600 to-blue-600 rounded-lg flex items-center justify-center font-bold text-sm text-white shadow-md">
            SSH
          </div>
          <div>
            <div className="font-semibold text-sm text-gray-900">跨平台 SSH 连接工具</div>
            <div className="text-xs text-gray-500">Cross-Platform SSH Manager</div>
          </div>
        </div>
        
        {/* 工具按钮 */}
        <div className="ml-auto">
          <button
            onClick={() => setIsDevToolsOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-lg font-medium transition-all shadow-sm"
          >
            <Wrench className="w-4 h-4" />
            <span className="text-sm">开发工具</span>
          </button>
        </div>
      </div>

      {/* 主内容区域 */}
      <div className="flex-1 flex overflow-hidden">
        {/* 左侧：资产列表 */}
        <div className="w-72 flex-shrink-0 shadow-sm">
          <AssetList 
            onConnect={handleConnect} 
            onAddClick={() => setIsAddDialogOpen(true)}
            assets={assets}
          />
        </div>

        {/* 中间：终端面板 */}
        <div className="flex-1 min-w-0">
          <TerminalPanel 
            connections={connections}
            onCloseConnection={handleCloseConnection}
          />
        </div>

        {/* 右侧：文件管理和服务器监控 */}
        <div className="w-80 flex-shrink-0 flex flex-col shadow-sm">
          {/* 文件管理 */}
          <div className="h-1/2">
            <FileManager />
          </div>
          
          {/* 服务器监控 */}
          <div className="h-1/2">
            <ServerMonitor />
          </div>
        </div>
      </div>

      {/* 对话框 */}
      <AddAssetDialog
        isOpen={isAddDialogOpen}
        onClose={() => setIsAddDialogOpen(false)}
        onAdd={handleAddAsset}
      />

      <DevToolsPanel
        isOpen={isDevToolsOpen}
        onClose={() => setIsDevToolsOpen(false)}
      />
    </div>
  );
}