import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * 合并 Tailwind CSS 类名
 * @param {...any} inputs - 要合并的类名
 * @returns {string} 合并后的类名字符串
 */
export function cn(...inputs) {
  return twMerge(clsx(inputs))
}
