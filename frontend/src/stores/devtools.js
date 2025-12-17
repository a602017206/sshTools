import { writable } from 'svelte/store';

// 工具注册表
const toolsRegistry = [];

/**
 * 注册新工具到工具集
 * @param {Object} toolConfig - 工具配置对象
 * @param {string} toolConfig.id - 唯一标识符
 * @param {string} toolConfig.name - 显示名称
 * @param {string} toolConfig.description - 工具描述
 * @param {string} toolConfig.icon - 图标（emoji或SVG）
 * @param {Component} toolConfig.component - Svelte组件
 * @param {string} [toolConfig.category] - 分类（可选）
 * @param {number} [toolConfig.order] - 排序权重（可选，默认0）
 */
export function registerTool(toolConfig) {
  // 验证必需字段
  if (!toolConfig.id || !toolConfig.name || !toolConfig.component) {
    console.error('工具注册失败：缺少必需字段 (id, name, component)', toolConfig);
    return;
  }

  // 检查是否已存在相同ID
  const existingIndex = toolsRegistry.findIndex(t => t.id === toolConfig.id);
  if (existingIndex !== -1) {
    console.warn(`工具 "${toolConfig.id}" 已存在，将被覆盖`);
    toolsRegistry[existingIndex] = toolConfig;
  } else {
    toolsRegistry.push(toolConfig);
  }
}

/**
 * 获取所有注册的工具（按order排序）
 * @returns {Array} 工具配置数组
 */
export function getRegisteredTools() {
  return [...toolsRegistry].sort((a, b) => (a.order || 0) - (b.order || 0));
}

/**
 * 根据ID获取工具
 * @param {string} id - 工具ID
 * @returns {Object|null} 工具配置对象
 */
export function getToolById(id) {
  return toolsRegistry.find(t => t.id === id) || null;
}

/**
 * 创建DevTools Store用于状态管理
 */
function createDevToolsStore() {
  const { subscribe, set, update } = writable({
    isOpen: false,           // 面板是否打开
    activeTool: null,        // 当前激活的工具ID
    position: 'right',       // 面板位置：'right' | 'bottom' | 'left'
    width: 500,              // 右侧/左侧面板宽度（px）
    height: 300,             // 底部面板高度（px）
  });

  return {
    subscribe,

    /**
     * 切换面板开关状态
     */
    toggle: () => update(state => ({
      ...state,
      isOpen: !state.isOpen
    })),

    /**
     * 打开面板
     */
    open: () => update(state => ({
      ...state,
      isOpen: true
    })),

    /**
     * 关闭面板
     */
    close: () => update(state => ({
      ...state,
      isOpen: false,
      activeTool: null  // 关闭时清空选中工具
    })),

    /**
     * 设置激活的工具
     * @param {string} toolId - 工具ID
     */
    setActiveTool: (toolId) => update(state => ({
      ...state,
      activeTool: toolId,
      isOpen: true  // 自动打开面板
    })),

    /**
     * 设置面板宽度
     * @param {number} width - 宽度（px）
     */
    setWidth: (width) => update(state => ({
      ...state,
      width: Math.max(300, Math.min(800, width))  // 限制范围
    })),

    /**
     * 设置面板高度
     * @param {number} height - 高度（px）
     */
    setHeight: (height) => update(state => ({
      ...state,
      height: Math.max(200, Math.min(600, height))  // 限制范围
    })),

    /**
     * 设置面板位置
     * @param {string} position - 位置 ('right' | 'bottom' | 'left')
     */
    setPosition: (position) => update(state => ({
      ...state,
      position
    })),

    /**
     * 重置到默认状态
     */
    reset: () => set({
      isOpen: false,
      activeTool: null,
      position: 'right',
      width: 500,
      height: 300,
    }),
  };
}

export const devToolsStore = createDevToolsStore();
