/**
 * Wails Go bindings - 此文件提供 Go 后端与前端 TypeScript 的桥接
 * 运行 'wails dev' 或 'wails generate module' 可重新生成
 */

/**
 * 检查是否在 Wails 环境中运行
 */
function isWailsEnv(): boolean {
  return !!(window as any).go?.main?.App;
}

/**
 * 返回问候语
 * @param name - 用户名
 * @returns 问候字符串
 */
export function Greet(arg1: string): Promise<string> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.Greet(arg1);
  }
  return Promise.resolve("Hello " + arg1 + "!");
}

/**
 * 获取当前操作系统
 * @returns 操作系统名称 (windows, darwin, linux 等)
 */
export function GetOS(): Promise<string> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.GetOS();
  }
  return Promise.resolve("browser");
}

/**
 * 最小化窗口
 */
export function Minimize(): Promise<void> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.Minimize();
  }
  return Promise.resolve();
}

/**
 * 最大化窗口
 */
export function Maximize(): Promise<void> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.Maximize();
  }
  return Promise.resolve();
}

/**
 * 取消窗口最大化状态
 */
export function Unmaximize(): Promise<void> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.Unmaximize();
  }
  return Promise.resolve();
}

/**
 * 关闭应用程序
 */
export function Close(): Promise<void> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.Close();
  }
  return Promise.resolve();
}

/**
 * 打开目录选择对话框
 * @returns 选中的目录路径，取消则返回空字符串
 */
export function SelectDirectory(): Promise<string> {
  if (isWailsEnv()) {
    return (window as any).go.main.App.SelectDirectory();
  }
  return Promise.resolve("");
}
