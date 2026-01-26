import { useState } from 'react';
import { X, Code, Hash, Lock, Calendar, Palette, Binary, FileJson, Copy } from 'lucide-react';

interface DevToolsPanelProps {
  isOpen: boolean;
  onClose: () => void;
}

export function DevToolsPanel({ isOpen, onClose }: DevToolsPanelProps) {
  const [activeToolId, setActiveToolId] = useState<string | null>(null);
  const [jsonInput, setJsonInput] = useState('');
  const [jsonOutput, setJsonOutput] = useState('');
  const [base64Input, setBase64Input] = useState('');
  const [base64Output, setBase64Output] = useState('');
  const [hashInput, setHashInput] = useState('');
  const [md5Output, setMd5Output] = useState('');
  const [generatedUUID, setGeneratedUUID] = useState('');
  const [copySuccess, setCopySuccess] = useState(false);

  if (!isOpen) return null;

  const tools = [
    { id: 'json', name: 'JSON 格式化', icon: FileJson, color: 'text-purple-600' },
    { id: 'base64', name: 'Base64 编解码', icon: Binary, color: 'text-blue-600' },
    { id: 'hash', name: 'Hash 计算', icon: Hash, color: 'text-green-600' },
    { id: 'timestamp', name: '时间戳转换', icon: Calendar, color: 'text-amber-600' },
    { id: 'color', name: '颜色转换', icon: Palette, color: 'text-pink-600' },
    { id: 'uuid', name: 'UUID 生成', icon: Code, color: 'text-indigo-600' },
  ];

  const formatJSON = () => {
    try {
      const parsed = JSON.parse(jsonInput);
      setJsonOutput(JSON.stringify(parsed, null, 2));
    } catch (error) {
      setJsonOutput('JSON 格式错误');
    }
  };

  const encodeBase64 = () => {
    try {
      setBase64Output(btoa(base64Input));
    } catch (error) {
      setBase64Output('编码失败');
    }
  };

  const decodeBase64 = () => {
    try {
      setBase64Output(atob(base64Input));
    } catch (error) {
      setBase64Output('解码失败');
    }
  };

  const calculateHash = async () => {
    // 简单的模拟，实际应用中使用 crypto API
    const encoder = new TextEncoder();
    const data = encoder.encode(hashInput);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    setMd5Output(hashHex);
  };

  const renderToolContent = () => {
    switch (activeToolId) {
      case 'json':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">输入 JSON</label>
              <textarea
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                placeholder='{"key": "value"}'
                className="w-full h-32 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
            </div>
            <button
              onClick={formatJSON}
              className="w-full px-4 py-2.5 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors"
            >
              格式化
            </button>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">输出</label>
              <textarea
                value={jsonOutput}
                readOnly
                className="w-full h-32 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono resize-none focus:outline-none"
              />
            </div>
          </div>
        );

      case 'base64':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">输入</label>
              <textarea
                value={base64Input}
                onChange={(e) => setBase64Input(e.target.value)}
                placeholder="输入要编码或解码的文本"
                className="w-full h-32 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={encodeBase64}
                className="px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
              >
                编码
              </button>
              <button
                onClick={decodeBase64}
                className="px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
              >
                解码
              </button>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">输出</label>
              <textarea
                value={base64Output}
                readOnly
                className="w-full h-32 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono resize-none focus:outline-none"
              />
            </div>
          </div>
        );

      case 'hash':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">输入文本</label>
              <textarea
                value={hashInput}
                onChange={(e) => setHashInput(e.target.value)}
                placeholder="输入要计算���希的文本"
                className="w-full h-32 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-green-500"
              />
            </div>
            <button
              onClick={calculateHash}
              className="w-full px-4 py-2.5 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors"
            >
              计算 SHA-256
            </button>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">SHA-256</label>
              <input
                value={md5Output}
                readOnly
                className="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none"
              />
            </div>
          </div>
        );

      case 'timestamp':
        return (
          <div className="space-y-4 text-center py-8">
            <div className="text-sm text-gray-500">当前时间戳</div>
            <div className="text-3xl font-bold text-gray-900">{Date.now()}</div>
            <div className="text-sm text-gray-500">{new Date().toLocaleString('zh-CN')}</div>
          </div>
        );

      case 'uuid':
        return (
          <div className="space-y-4">
            <button
              onClick={() => {
                const uuid = crypto.randomUUID();
                setGeneratedUUID(uuid);
                setCopySuccess(false);
              }}
              className="w-full px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-medium transition-colors"
            >
              生成 UUID
            </button>
            
            {generatedUUID && (
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">生成的 UUID</label>
                <div className="flex gap-2">
                  <input
                    value={generatedUUID}
                    readOnly
                    className="flex-1 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none"
                  />
                  <button
                    onClick={() => {
                      // 使用传统的 textarea 复制方法
                      const textarea = document.createElement('textarea');
                      textarea.value = generatedUUID;
                      textarea.style.position = 'fixed';
                      textarea.style.opacity = '0';
                      document.body.appendChild(textarea);
                      textarea.select();
                      try {
                        document.execCommand('copy');
                        setCopySuccess(true);
                        setTimeout(() => setCopySuccess(false), 2000);
                      } catch (err) {
                        console.error('复制失败:', err);
                      }
                      document.body.removeChild(textarea);
                    }}
                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                    title="复制到剪贴板"
                  >
                    <Copy className="w-4 h-4 text-gray-700" />
                  </button>
                </div>
                {copySuccess && (
                  <div className="text-xs text-green-600">已复制到剪贴板！</div>
                )}
              </div>
            )}
          </div>
        );

      default:
        return (
          <div className="text-center py-12 text-gray-500">
            选择一个工具开始使用
          </div>
        );
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16">
      {/* 背景遮罩 */}
      <div 
        className="absolute inset-0 bg-black/20 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* 面板 */}
      <div className="relative w-full max-w-4xl bg-white rounded-2xl shadow-2xl m-4 max-h-[80vh] overflow-hidden flex">
        {/* 左侧工具列表 */}
        <div className="w-64 bg-gray-50 border-r border-gray-200 p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-gray-900">开发工具</h3>
            <button
              onClick={onClose}
              className="p-1.5 hover:bg-gray-200 rounded-lg transition-colors"
            >
              <X className="w-4 h-4 text-gray-500" />
            </button>
          </div>
          
          <div className="space-y-1">
            {tools.map((tool) => (
              <button
                key={tool.id}
                onClick={() => setActiveToolId(tool.id)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all ${
                  activeToolId === tool.id
                    ? 'bg-white shadow-sm border border-gray-200'
                    : 'hover:bg-gray-100'
                }`}
              >
                <tool.icon className={`w-4 h-4 ${tool.color}`} />
                <span className={`text-sm font-medium ${
                  activeToolId === tool.id ? 'text-gray-900' : 'text-gray-700'
                }`}>
                  {tool.name}
                </span>
              </button>
            ))}
          </div>
        </div>

        {/* 右侧工具内容 */}
        <div className="flex-1 p-6 overflow-y-auto">
          <h2 className="text-lg font-semibold text-gray-900 mb-6">
            {tools.find(t => t.id === activeToolId)?.name || '选择工具'}
          </h2>
          {renderToolContent()}
        </div>
      </div>
    </div>
  );
}