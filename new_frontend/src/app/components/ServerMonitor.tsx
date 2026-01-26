import { useState, useEffect } from 'react';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Cpu, HardDrive, Activity, Wifi } from 'lucide-react';

export function ServerMonitor() {
  const [cpuData, setCpuData] = useState<Array<{ time: string; value: number }>>([]);
  const [memoryData, setMemoryData] = useState<Array<{ time: string; value: number }>>([]);
  const [currentStats, setCurrentStats] = useState({
    cpu: 45,
    memory: 62,
    disk: 78,
    network: { in: 125.5, out: 89.3 }
  });

  // 模拟实时数据更新
  useEffect(() => {
    // 初始化数据
    const initData = Array.from({ length: 10 }, (_, i) => ({
      time: `00:00:${i.toString().padStart(2, '0')}`,
      value: Math.floor(Math.random() * 40) + 30
    }));
    setCpuData(initData);
    setMemoryData(initData);

    const generateDataPoint = (prev: number) => {
      const change = (Math.random() - 0.5) * 10;
      const newValue = Math.max(0, Math.min(100, prev + change));
      return Math.round(newValue);
    };

    const interval = setInterval(() => {
      const now = new Date();
      const timeStr = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`;
      
      setCpuData(prev => {
        const newData = [...prev, { time: timeStr, value: generateDataPoint(prev[prev.length - 1]?.value || 45) }];
        return newData.slice(-20);
      });

      setMemoryData(prev => {
        const newData = [...prev, { time: timeStr, value: generateDataPoint(prev[prev.length - 1]?.value || 62) }];
        return newData.slice(-20);
      });

      setCurrentStats(prev => ({
        cpu: generateDataPoint(prev.cpu),
        memory: generateDataPoint(prev.memory),
        disk: prev.disk,
        network: {
          in: Math.max(0, prev.network.in + (Math.random() - 0.5) * 20),
          out: Math.max(0, prev.network.out + (Math.random() - 0.5) * 15)
        }
      }));
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  const getStatusColor = (value: number) => {
    if (value < 50) return '#10b981';
    if (value < 80) return '#f59e0b';
    return '#ef4444';
  };

  return (
    <div className="h-full flex flex-col bg-white overflow-y-auto">
      {/* 头部 */}
      <div className="p-3 border-b border-gray-200">
        <h3 className="text-sm font-semibold text-gray-900">服务器监控</h3>
      </div>

      <div className="p-3 space-y-3">
        {/* CPU 使用率 */}
        <div className="bg-gradient-to-br from-purple-50 to-blue-50 rounded-xl p-3 shadow-sm border border-purple-100">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <div className="p-1.5 bg-purple-100 rounded-lg">
                <Cpu className="w-3.5 h-3.5 text-purple-600" />
              </div>
              <span className="text-xs font-semibold text-gray-900">CPU</span>
            </div>
            <span className="text-xs font-bold px-2 py-1 rounded-lg bg-white" style={{ color: getStatusColor(currentStats.cpu) }}>
              {currentStats.cpu}%
            </span>
          </div>
          <div className="h-[80px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={cpuData}>
                <defs>
                  <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#9333ea" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#9333ea" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <Area type="monotone" dataKey="value" stroke="#9333ea" fill="url(#cpuGradient)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* 内存使用率 */}
        <div className="bg-gradient-to-br from-emerald-50 to-teal-50 rounded-xl p-3 shadow-sm border border-emerald-100">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <div className="p-1.5 bg-emerald-100 rounded-lg">
                <Activity className="w-3.5 h-3.5 text-emerald-600" />
              </div>
              <span className="text-xs font-semibold text-gray-900">内存</span>
            </div>
            <span className="text-xs font-bold px-2 py-1 rounded-lg bg-white" style={{ color: getStatusColor(currentStats.memory) }}>
              {currentStats.memory}%
            </span>
          </div>
          <div className="h-[80px]">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={memoryData}>
                <defs>
                  <linearGradient id="memoryGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <Area type="monotone" dataKey="value" stroke="#10b981" fill="url(#memoryGradient)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* 磁盘使用 */}
        <div className="bg-gradient-to-br from-amber-50 to-orange-50 rounded-xl p-3 shadow-sm border border-amber-100">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <div className="p-1.5 bg-amber-100 rounded-lg">
                <HardDrive className="w-3.5 h-3.5 text-amber-600" />
              </div>
              <span className="text-xs font-semibold text-gray-900">磁盘</span>
            </div>
            <span className="text-xs font-bold px-2 py-1 rounded-lg bg-white" style={{ color: getStatusColor(currentStats.disk) }}>
              {currentStats.disk}%
            </span>
          </div>
          <div className="w-full bg-gray-100 rounded-full h-2 overflow-hidden">
            <div
              className="h-2 rounded-full transition-all duration-300"
              style={{ 
                width: `${currentStats.disk}%`,
                backgroundColor: getStatusColor(currentStats.disk)
              }}
            />
          </div>
          <div className="flex justify-between mt-2 text-[10px] text-gray-600">
            <span>已用 156 GB</span>
            <span>总计 200 GB</span>
          </div>
        </div>

        {/* 网络流量 */}
        <div className="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-xl p-3 shadow-sm border border-blue-100">
          <div className="flex items-center gap-2 mb-3">
            <div className="p-1.5 bg-blue-100 rounded-lg">
              <Wifi className="w-3.5 h-3.5 text-blue-600" />
            </div>
            <span className="text-xs font-semibold text-gray-900">网络流量</span>
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between bg-white rounded-lg p-2">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-500" />
                <span className="text-xs text-gray-700 font-medium">入站</span>
              </div>
              <span className="text-xs font-mono font-bold text-gray-900">
                {currentStats.network.in.toFixed(1)} MB/s
              </span>
            </div>
            <div className="flex items-center justify-between bg-white rounded-lg p-2">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-rose-500" />
                <span className="text-xs text-gray-700 font-medium">出站</span>
              </div>
              <span className="text-xs font-mono font-bold text-gray-900">
                {currentStats.network.out.toFixed(1)} MB/s
              </span>
            </div>
          </div>
        </div>

        {/* 系统信息 */}
        <div className="bg-gray-50 rounded-xl p-3 shadow-sm border border-gray-200">
          <div className="text-xs font-semibold text-gray-900 mb-2">系统信息</div>
          <div className="space-y-1.5 text-[10px]">
            <div className="flex justify-between py-1">
              <span className="text-gray-600">操作系统</span>
              <span className="text-gray-900 font-medium">Ubuntu 22.04 LTS</span>
            </div>
            <div className="flex justify-between py-1">
              <span className="text-gray-600">内核版本</span>
              <span className="text-gray-900 font-medium">5.15.0-91-generic</span>
            </div>
            <div className="flex justify-between py-1">
              <span className="text-gray-600">运行时间</span>
              <span className="text-gray-900 font-medium">15天 7小时 32分</span>
            </div>
            <div className="flex justify-between py-1">
              <span className="text-gray-600">进程数</span>
              <span className="text-gray-900 font-medium">187</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}