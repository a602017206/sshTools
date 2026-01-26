/**
 * Dialog utilities for Wails application
 * Provides cross-platform dialog methods that work in both browser and Wails client
 */

import { ShowMessageDialog, ShowErrorDialog, ShowQuestionDialog } from '../../wailsjs/go/main/App.js';

/**
 * Show an alert/message dialog
 * @param {string} message - The message to display
 * @param {string} title - Optional title (default: "提示")
 */
export async function showAlert(message, title = "提示") {
  try {
    // Use Wails Go backend method for native dialogs
    await ShowMessageDialog(title, message);
  } catch (error) {
    // Fallback to browser alert if Wails is not available (dev mode in browser)
    console.warn('Wails backend not available, using browser alert:', error);
    alert(message);
  }
}

/**
 * Show an error dialog
 * @param {string} message - The error message to display
 * @param {string} title - Optional title (default: "错误")
 */
export async function showError(message, title = "错误") {
  try {
    // Use Wails Go backend method for native dialogs
    await ShowErrorDialog(title, message);
  } catch (error) {
    console.warn('Wails backend not available, using browser alert:', error);
    alert(`${title}: ${message}`);
  }
}

/**
 * Show a confirmation dialog
 * @param {string} message - The question to ask
 * @param {string} title - Optional title (default: "确认")
 * @returns {Promise<boolean>} - true if user confirms, false otherwise
 */
export async function showConfirm(message, title = "确认") {
  try {
    // Use Wails Go backend method for native dialogs
    const result = await ShowQuestionDialog(title, message);
    return result;
  } catch (error) {
    // Fallback to browser confirm if Wails is not available
    console.warn('Wails backend not available, using browser confirm:', error);
    return confirm(message);
  }
}

/**
 * Show a prompt dialog (for password input)
 * Note: Wails doesn't have a built-in prompt dialog, so we fall back to browser prompt
 * For password input, consider creating a custom Svelte modal component instead
 *
 * @param {string} message - The prompt message
 * @param {string} defaultValue - Default value
 * @returns {Promise<string|null>} - The user's input or null if cancelled
 */
export async function showPrompt(message, defaultValue = '') {
  // For now, fall back to browser prompt
  // TODO: Create a custom Svelte modal component for better UX
  console.warn('showPrompt using browser prompt - consider implementing custom modal');
  return prompt(message, defaultValue);
}
