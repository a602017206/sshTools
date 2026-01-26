import { useState, useEffect, useRef } from 'react';
import { X, Minimize2, Maximize2, Copy } from 'lucide-react';
import type { ServerAsset } from './AssetList';

interface TerminalTab {
  id: string;
  asset: ServerAsset;
  active: boolean;
}

interface TerminalPanelProps {
  connections: ServerAsset[];
  onCloseConnection: (id: string) => void;
}

export function TerminalPanel({ connections, onCloseConnection }: TerminalPanelProps) {
  const [activeTab, setActiveTab] = useState<string | null>(null);
  const [tabs, setTabs] = useState<TerminalTab[]>([]);
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const newTabs = connections.map(conn => ({
      id: conn.id,
      asset: conn,
      active: conn.id === activeTab
    }));
    setTabs(newTabs);
    
    if (connections.length > 0 && !activeTab) {
      setActiveTab(connections[0].id);
    }
  }, [connections, activeTab]);

  const handleCloseTab = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    onCloseConnection(id);
    if (activeTab === id) {
      const remainingTabs = tabs.filter(t => t.id !== id);
      setActiveTab(remainingTabs.length > 0 ? remainingTabs[0].id : null);
    }
  };

  const activeConnection = connections.find(c => c.id === activeTab);

  // 模拟终端输出
  const mockTerminalOutput = activeConnection ? [
    `Connecting to ${activeConnection.username}@${activeConnection.host}:${activeConnection.port}...`,
    `Welcome to Ubuntu 22.04.3 LTS (GNU/Linux 5.15.0-91-generic x86_64)`,
    ``,
    ` * Documentation:  https://help.ubuntu.com`,
    ` * Management:     https://landscape.canonical.com`,
    ` * Support:        https://ubuntu.com/advantage`,
    ``,
    `Last login: ${new Date().toLocaleString('zh-CN')} from 192.168.1.100`,
    `${activeConnection.username}@${activeConnection.name}:~$ _`
  ] : [];

  return (
    <div className="h-full flex flex-col bg-gray-50">
      {/* 标签栏 */}
      <div className="flex items-center bg-white border-b border-gray-200 overflow-x-auto">
        {tabs.length === 0 ? (
          <div className="px-4 py-2.5 text-sm text-gray-500">没有活动连接</div>
        ) : (
          tabs.map(tab => (
            <div
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`group flex items-center gap-2 px-4 py-2.5 border-r border-gray-200 cursor-pointer transition-all min-w-[180px] ${
                activeTab === tab.id 
                  ? 'bg-white text-gray-900 border-b-2 border-b-purple-600' 
                  : 'bg-gray-50 text-gray-600 hover:bg-gray-100'
              }`}
            >
              <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
                tab.asset.status === 'online' ? 'bg-green-500' : 'bg-gray-400'
              }`} />
              <span className="text-sm font-medium truncate flex-1">{tab.asset.name}</span>
              <button
                onClick={(e) => handleCloseTab(tab.id, e)}
                className="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-gray-200 rounded transition-opacity"
              >
                <X className="w-3 h-3" />
              </button>
            </div>
          ))
        )}
      </div>

      {/* 终端内容 */}
      {activeConnection ? (
        <div className="flex-1 flex flex-col">
          {/* 工具栏 */}
          <div className="flex items-center justify-between px-4 py-2.5 bg-white border-b border-gray-200">
            <div className="text-sm text-gray-600 font-mono">
              {activeConnection.username}@{activeConnection.host}:{activeConnection.port}
            </div>
            <div className="flex items-center gap-1">
              <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                <Copy className="w-4 h-4 text-gray-600" />
              </button>
              <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                <Minimize2 className="w-4 h-4 text-gray-600" />
              </button>
              <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors">
                <Maximize2 className="w-4 h-4 text-gray-600" />
              </button>
            </div>
          </div>

          {/* 终端窗口 */}
          <div 
            ref={terminalRef}
            className="flex-1 p-6 overflow-y-auto font-mono text-sm bg-white shadow-inner"
          >
            {mockTerminalOutput.map((line, index) => (
              <div key={index} className="text-gray-800 leading-relaxed">
                {line}
              </div>
            ))}
          </div>

          {/* 输入区域 */}
          <div className="p-4 bg-white border-t border-gray-200 shadow-sm">
            <div className="flex items-center gap-2 font-mono text-sm bg-gray-50 rounded-lg px-4 py-3">
              <span className="text-purple-600 font-semibold">{activeConnection.username}@{activeConnection.name}:~$</span>
              <input
                type="text"
                className="flex-1 bg-transparent text-gray-900 outline-none"
                placeholder="输入命令..."
                autoFocus
              />
            </div>
          </div>
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center text-gray-500 bg-white">
          <div className="text-center">
            <div className="text-lg font-medium mb-2 text-gray-700">未选择连接</div>
            <div className="text-sm text-gray-500">从左侧资产列表选择一个服务器开始连接</div>
          </div>
        </div>
      )}
    </div>
  );
}