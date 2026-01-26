import { useState } from 'react';
import { X, Server, Database, Container } from 'lucide-react';

export type AssetType = 'ssh' | 'database' | 'docker';

interface AddAssetDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onAdd: (asset: any) => void;
}

export function AddAssetDialog({ isOpen, onClose, onAdd }: AddAssetDialogProps) {
  const [assetType, setAssetType] = useState<AssetType>('ssh');
  const [formData, setFormData] = useState({
    name: '',
    host: '',
    port: '',
    username: '',
    password: '',
    group: '',
    dbType: 'mysql',
    database: '',
  });

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newAsset = {
      id: Date.now().toString(),
      type: assetType,
      name: formData.name,
      host: formData.host,
      port: parseInt(formData.port),
      username: formData.username,
      group: formData.group || '默认分组',
      status: 'online' as const,
      ...(assetType === 'database' && { 
        dbType: formData.dbType,
        database: formData.database 
      }),
    };
    onAdd(newAsset);
    onClose();
    // 重置表单
    setFormData({
      name: '',
      host: '',
      port: '',
      username: '',
      password: '',
      group: '',
      dbType: 'mysql',
      database: '',
    });
  };

  const getDefaultPort = () => {
    switch (assetType) {
      case 'ssh': return '22';
      case 'docker': return '2375';
      case 'database': 
        switch (formData.dbType) {
          case 'mysql': return '3306';
          case 'postgresql': return '5432';
          case 'mongodb': return '27017';
          case 'redis': return '6379';
          default: return '';
        }
      default: return '';
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* 背景遮罩 */}
      <div 
        className="absolute inset-0 bg-black/20 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* 对话框 */}
      <div className="relative w-full max-w-lg bg-white rounded-2xl shadow-2xl p-6 m-4 max-h-[90vh] overflow-y-auto">
        {/* 头部 */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">添加连接</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* 连接类型选择 */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-3">连接类型</label>
          <div className="grid grid-cols-3 gap-3">
            <button
              type="button"
              onClick={() => setAssetType('ssh')}
              className={`p-4 rounded-xl border-2 transition-all ${
                assetType === 'ssh'
                  ? 'border-purple-600 bg-purple-50 shadow-sm'
                  : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
              }`}
            >
              <Server className={`w-6 h-6 mx-auto mb-2 ${
                assetType === 'ssh' ? 'text-purple-600' : 'text-gray-600'
              }`} />
              <div className={`text-sm font-medium ${
                assetType === 'ssh' ? 'text-purple-900' : 'text-gray-700'
              }`}>SSH</div>
            </button>

            <button
              type="button"
              onClick={() => setAssetType('database')}
              className={`p-4 rounded-xl border-2 transition-all ${
                assetType === 'database'
                  ? 'border-purple-600 bg-purple-50 shadow-sm'
                  : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
              }`}
            >
              <Database className={`w-6 h-6 mx-auto mb-2 ${
                assetType === 'database' ? 'text-purple-600' : 'text-gray-600'
              }`} />
              <div className={`text-sm font-medium ${
                assetType === 'database' ? 'text-purple-900' : 'text-gray-700'
              }`}>数据库</div>
            </button>

            <button
              type="button"
              onClick={() => setAssetType('docker')}
              className={`p-4 rounded-xl border-2 transition-all ${
                assetType === 'docker'
                  ? 'border-purple-600 bg-purple-50 shadow-sm'
                  : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
              }`}
            >
              <Container className={`w-6 h-6 mx-auto mb-2 ${
                assetType === 'docker' ? 'text-purple-600' : 'text-gray-600'
              }`} />
              <div className={`text-sm font-medium ${
                assetType === 'docker' ? 'text-purple-900' : 'text-gray-700'
              }`}>Docker</div>
            </button>
          </div>
        </div>

        {/* 表单 */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* 连接名称 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              连接名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="例如：生产服务器-01"
              className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
            />
          </div>

          {/* 数据库类型（仅数据库连接显示） */}
          {assetType === 'database' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                数据库类型 <span className="text-red-500">*</span>
              </label>
              <select
                value={formData.dbType}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  dbType: e.target.value,
                  port: '' // 重置端口，让用户选择新的默认值
                })}
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              >
                <option value="mysql">MySQL</option>
                <option value="postgresql">PostgreSQL</option>
                <option value="mongodb">MongoDB</option>
                <option value="redis">Redis</option>
              </select>
            </div>
          )}

          {/* 主机地址 */}
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                主机地址 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                required
                value={formData.host}
                onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                placeholder="192.168.1.10 或 example.com"
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                端口 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                required
                value={formData.port}
                onChange={(e) => setFormData({ ...formData, port: e.target.value })}
                placeholder={getDefaultPort()}
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            </div>
          </div>

          {/* 用户名和密码 */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                用户名 <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                required
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                placeholder="root"
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                密码
              </label>
              <input
                type="password"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                placeholder="••••••••"
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            </div>
          </div>

          {/* 数据库名（仅数据库连接显示） */}
          {assetType === 'database' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                数据库名
              </label>
              <input
                type="text"
                value={formData.database}
                onChange={(e) => setFormData({ ...formData, database: e.target.value })}
                placeholder="例如：production_db"
                className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
              />
            </div>
          )}

          {/* 分组 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              分组
            </label>
            <input
              type="text"
              value={formData.group}
              onChange={(e) => setFormData({ ...formData, group: e.target.value })}
              placeholder="例如：生产环境"
              className="w-full px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
            />
          </div>

          {/* 按钮 */}
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg font-medium transition-colors"
            >
              取消
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2.5 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-lg font-medium transition-all shadow-sm"
            >
              添加连接
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
