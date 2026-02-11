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

export const formatTime = (isoString) => {
  if (!isoString) return "";
  const date = new Date(isoString);
  const now = new Date();
  const diff = now - date;

  if (diff < 60000) {
    return "刚刚";
  } else if (diff < 3600000) {
    return `${Math.floor(diff / 60000)} 分钟前`;
  } else if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)} 小时前`;
  } else {
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
  }
};