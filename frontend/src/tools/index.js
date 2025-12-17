/**
 * 开发工具集 - 工具注册中心
 *
 * 在此文件中注册所有可用的开发工具。
 * 每个工具都是一个独立的Svelte组件，通过registerTool函数注册。
 *
 * 添加新工具的步骤：
 * 1. 在 components/tools/ 目录下创建新的工具组件
 * 2. 导入组件
 * 3. 调用 registerTool() 注册工具
 */

import { registerTool } from '../stores/devtools.js';
import JsonFormatter from '../components/tools/JsonFormatter.svelte';

// ====================================
// 注册JSON格式化工具
// ====================================
registerTool({
  id: 'json-formatter',
  name: 'JSON 格式化',
  description: 'JSON格式化、验证和语法高亮',
  icon: '{ }',
  component: JsonFormatter,
  category: 'formatter',
  order: 1
});

// ====================================
// 未来工具扩展示例（已注释）
// ====================================

// 示例1: Base64编解码工具
// import Base64Tool from '../components/tools/Base64Tool.svelte';
// registerTool({
//   id: 'base64',
//   name: 'Base64',
//   description: 'Base64编码和解码',
//   icon: '🔐',
//   component: Base64Tool,
//   category: 'encoder',
//   order: 2
// });

// 示例2: URL编解码工具
// import UrlTool from '../components/tools/UrlTool.svelte';
// registerTool({
//   id: 'url-encoder',
//   name: 'URL 编解码',
//   description: 'URL编码和解码',
//   icon: '🔗',
//   component: UrlTool,
//   category: 'encoder',
//   order: 3
// });

// 示例3: 时间戳转换工具
// import TimestampTool from '../components/tools/TimestampTool.svelte';
// registerTool({
//   id: 'timestamp',
//   name: '时间戳转换',
//   description: 'Unix时间戳与日期时间互转',
//   icon: '⏰',
//   component: TimestampTool,
//   category: 'converter',
//   order: 4
// });

// 示例4: Hash计算器
// import HashTool from '../components/tools/HashTool.svelte';
// registerTool({
//   id: 'hash',
//   name: 'Hash 计算',
//   description: 'MD5/SHA256/SHA512哈希计算',
//   icon: '🔑',
//   component: HashTool,
//   category: 'crypto',
//   order: 5
// });

// 示例5: UUID生成器
// import UuidTool from '../components/tools/UuidTool.svelte';
// registerTool({
//   id: 'uuid',
//   name: 'UUID 生成器',
//   description: '生成各种版本的UUID',
//   icon: '🆔',
//   component: UuidTool,
//   category: 'generator',
//   order: 6
// });

// 示例6: 正则表达式测试器
// import RegexTool from '../components/tools/RegexTool.svelte';
// registerTool({
//   id: 'regex',
//   name: '正则测试',
//   description: '正则表达式测试和匹配',
//   icon: '🔍',
//   component: RegexTool,
//   category: 'tester',
//   order: 7
// });

// 示例7: 颜色转换器
// import ColorTool from '../components/tools/ColorTool.svelte';
// registerTool({
//   id: 'color',
//   name: '颜色转换',
//   description: 'HEX/RGB/HSL颜色转换',
//   icon: '🎨',
//   component: ColorTool,
//   category: 'converter',
//   order: 8
// });

/**
 * 初始化开发工具集
 * 在应用启动时调用
 */
export function initializeDevTools() {
  console.log('开发工具集初始化完成');
  console.log(`已注册 ${getRegisteredTools().length} 个工具`);
}

/**
 * 获取已注册的工具列表（用于调试）
 */
import { getRegisteredTools } from '../stores/devtools.js';
export { getRegisteredTools };
