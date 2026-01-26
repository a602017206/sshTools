import { useState } from 'react';
import { Folder, File, ChevronRight, ChevronDown, Download, Upload, RefreshCw, Home, Search, X } from 'lucide-react';

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
  const [isPathInputOpen, setIsPathInputOpen] = useState(false);
  const [pathInput, setPathInput] = useState('');
  const [isEditingPath, setIsEditingPath] = useState(false);
  const [editPathValue, setEditPathValue] = useState('');

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

  const handlePathJump = () => {
    if (pathInput.trim()) {
      setCurrentPath(pathInput.trim());
      setIsPathInputOpen(false);
      setPathInput('');
    }
  };

  const handleGoHome = () => {
    setCurrentPath('/home/user');
  };

  const handleStartEditPath = () => {
    setEditPathValue(currentPath);
    setIsEditingPath(true);
  };

  const handleSaveEditPath = () => {
    if (editPathValue.trim()) {
      setCurrentPath(editPathValue.trim());
    }
    setIsEditingPath(false);
  };

  const handleCancelEditPath = () => {
    setIsEditingPath(false);
    setEditPathValue('');
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
          <div className="flex items-center gap-2">
            <div className="flex-1 flex items-center gap-1 text-xs bg-gray-50 rounded-lg px-3 py-2">
              <button
                  onClick={handleGoHome}
                  className="p-0.5 hover:bg-gray-200 rounded transition-colors"
                  title="回到根目录"
              >
                <Home className="w-3 h-3 text-gray-600" />
              </button>
              <span className="text-gray-400">/</span>
              {isEditingPath ? (
                  <input
                      type="text"
                      value={editPathValue}
                      onChange={(e) => setEditPathValue(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          handleSaveEditPath();
                        } else if (e.key === 'Escape') {
                          handleCancelEditPath();
                        }
                      }}
                      onBlur={handleSaveEditPath}
                      className="flex-1 bg-white border border-purple-300 rounded px-2 py-1 text-purple-600 font-medium focus:outline-none focus:ring-2 focus:ring-purple-500"
                      autoFocus
                  />
              ) : (
                  <span
                      onClick={handleStartEditPath}
                      className="text-purple-600 font-medium cursor-text hover:bg-purple-100 px-2 py-1 rounded transition-colors"
                      title="点击编辑路径"
                  >
                {currentPath.split('/').filter(Boolean).join(' / ')}
              </span>
              )}
            </div>

            {/* 路径搜索按钮 */}
            <button
                onClick={() => setIsPathInputOpen(!isPathInputOpen)}
                className="p-2 hover:bg-purple-100 rounded-lg transition-colors"
                title="跳转到指定目录"
            >
              <Search className="w-3.5 h-3.5 text-purple-600" />
            </button>
          </div>

          {/* 路径输入框 */}
          {isPathInputOpen && (
              <div className="mt-2 flex items-center gap-2 animate-in fade-in slide-in-from-top-2 duration-200">
                <input
                    type="text"
                    value={pathInput}
                    onChange={(e) => setPathInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handlePathJump();
                      } else if (e.key === 'Escape') {
                        setIsPathInputOpen(false);
                        setPathInput('');
                      }
                    }}
                    placeholder="输入路径，如：/var/www/html"
                    className="flex-1 px-3 py-2 text-xs bg-white border border-purple-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                    autoFocus
                />
                <button
                    onClick={handlePathJump}
                    className="px-3 py-2 bg-purple-600 hover:bg-purple-700 text-white text-xs rounded-lg transition-colors font-medium"
                >
                  跳转
                </button>
                <button
                    onClick={() => {
                      setIsPathInputOpen(false);
                      setPathInput('');
                    }}
                    className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                    title="取消"
                >
                  <X className="w-3.5 h-3.5 text-gray-600" />
                </button>
              </div>
          )}
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