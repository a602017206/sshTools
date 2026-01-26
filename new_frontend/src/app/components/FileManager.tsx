import { useState } from 'react';
import { Folder, File, ChevronRight, ChevronDown, Download, Upload, RefreshCw, Home } from 'lucide-react';

interface FileItem {
  name: string;
  type: 'file' | 'directory';
  size?: string;
  modified?: string;
  permissions?: string;
}

export function FileManager() {
  const [currentPath, setCurrentPath] = useState('/home/user');
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set(['/home/user']));

  // 模拟文件数据
  const files: FileItem[] = [
    { name: 'Documents', type: 'directory', modified: '2024-01-20 10:30', permissions: 'drwxr-xr-x' },
    { name: 'Downloads', type: 'directory', modified: '2024-01-20 09:15', permissions: 'drwxr-xr-x' },
    { name: 'Projects', type: 'directory', modified: '2024-01-22 14:20', permissions: 'drwxr-xr-x' },
    { name: 'config.json', type: 'file', size: '2.4 KB', modified: '2024-01-22 11:45', permissions: '-rw-r--r--' },
    { name: 'deploy.sh', type: 'file', size: '1.8 KB', modified: '2024-01-21 16:30', permissions: '-rwxr-xr-x' },
    { name: 'README.md', type: 'file', size: '4.2 KB', modified: '2024-01-20 08:00', permissions: '-rw-r--r--' },
    { name: 'package.json', type: 'file', size: '1.2 KB', modified: '2024-01-19 15:20', permissions: '-rw-r--r--' },
  ];

  const toggleDirectory = (dirName: string) => {
    const newExpanded = new Set(expandedDirs);
    const fullPath = `${currentPath}/${dirName}`;
    if (newExpanded.has(fullPath)) {
      newExpanded.delete(fullPath);
    } else {
      newExpanded.add(fullPath);
    }
    setExpandedDirs(newExpanded);
  };

  return (
    <div className="h-full flex flex-col bg-white border-b border-gray-200">
      {/* 头部工具栏 */}
      <div className="p-3 border-b border-gray-200">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-semibold text-gray-900">文件管理</h3>
          <div className="flex items-center gap-1">
            <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors" title="刷新">
              <RefreshCw className="w-3.5 h-3.5 text-gray-600" />
            </button>
            <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors" title="上传">
              <Upload className="w-3.5 h-3.5 text-gray-600" />
            </button>
            <button className="p-1.5 hover:bg-gray-100 rounded-lg transition-colors" title="下载">
              <Download className="w-3.5 h-3.5 text-gray-600" />
            </button>
          </div>
        </div>
        
        {/* 路径导航 */}
        <div className="flex items-center gap-1 text-xs bg-gray-50 rounded-lg px-3 py-2">
          <button className="p-0.5 hover:bg-gray-200 rounded transition-colors">
            <Home className="w-3 h-3 text-gray-600" />
          </button>
          <span className="text-gray-400">/</span>
          <span className="text-purple-600 font-medium">{currentPath.split('/').filter(Boolean).join(' / ')}</span>
        </div>
      </div>

      {/* 文件列表 */}
      <div className="flex-1 overflow-y-auto text-xs">
        {files.map((file, index) => (
          <div
            key={index}
            className="group flex items-center gap-2 px-3 py-2 hover:bg-purple-50 cursor-pointer transition-colors mx-2 my-0.5 rounded-lg"
            onClick={() => file.type === 'directory' && toggleDirectory(file.name)}
          >
            {file.type === 'directory' ? (
              <>
                {expandedDirs.has(`${currentPath}/${file.name}`) ? (
                  <ChevronDown className="w-3 h-3 text-gray-500 flex-shrink-0" />
                ) : (
                  <ChevronRight className="w-3 h-3 text-gray-500 flex-shrink-0" />
                )}
                <Folder className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />
              </>
            ) : (
              <>
                <div className="w-3" />
                <File className="w-3.5 h-3.5 text-blue-500 flex-shrink-0" />
              </>
            )}
            <div className="flex-1 min-w-0">
              <div className="text-gray-900 font-medium truncate">{file.name}</div>
              {file.type === 'file' && (
                <div className="text-gray-500 flex items-center gap-2">
                  <span>{file.size}</span>
                  <span>•</span>
                  <span>{file.modified}</span>
                </div>
              )}
            </div>
            <span className="text-gray-400 font-mono text-[10px]">{file.permissions}</span>
          </div>
        ))}
      </div>
    </div>
  );
}