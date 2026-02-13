<template>
  <div class="min-h-screen flex">
    <!-- 侧边栏 -->
    <aside
      :class="[
        'fixed inset-y-0 left-0 z-50 w-64 bg-sidebar border-r border-sidebar-border transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full',
      ]"
    >
      <!-- Logo -->
      <div class="h-16 flex items-center px-6 border-b border-sidebar-border">
        <Settings class="h-6 w-6 text-sidebar-primary mr-2" />
        <span class="text-lg font-semibold text-sidebar-foreground">后台管理</span>
      </div>

      <!-- 导航菜单 -->
      <nav class="p-4 space-y-1">
        <router-link
          v-for="item in navigation"
          :key="item.path"
          :to="item.path"
          :class="[
            'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
            isActive(item.path)
              ? 'bg-sidebar-primary text-sidebar-primary-foreground'
              : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          ]"
        >
          <component :is="item.icon" class="h-5 w-5" />
          {{ item.name }}
        </router-link>
      </nav>

      <!-- 底部信息 -->
      <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-sidebar-border">
        <div class="flex items-center gap-3 text-sm text-sidebar-muted-foreground">
          <div class="w-8 h-8 rounded-full bg-sidebar-accent flex items-center justify-center">
            <User class="h-4 w-4" />
          </div>
          <div>
            <p class="font-medium text-sidebar-foreground">管理员</p>
            <p class="text-xs">admin@example.com</p>
          </div>
        </div>
      </div>
    </aside>

    <!-- 遮罩层（移动端） -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-40 bg-black/50 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- 顶部导航栏 -->
      <header class="h-16 flex items-center justify-between px-4 sm:px-6 border-b bg-card">
        <div class="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            class="lg:hidden"
            @click="sidebarOpen = !sidebarOpen"
          >
            <Menu class="h-5 w-5" />
          </Button>
          <h1 class="text-lg font-semibold">{{ pageTitle }}</h1>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="ghost" size="icon" @click="toggleTheme">
            <Sun v-if="isDark" class="h-5 w-5" />
            <Moon v-else class="h-5 w-5" />
          </Button>
        </div>
      </header>

      <!-- 页面内容 -->
      <main class="flex-1 overflow-auto p-4 sm:p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import {
  Settings,
  Building2,
  Bot,
  Wrench,
  User,
  Menu,
  Sun,
  Moon,
  LayoutDashboard,
  Server,
} from "lucide-vue-next";

const route = useRoute();
const router = useRouter();

// 侧边栏状态
const sidebarOpen = ref(false);

// 主题状态
const isDark = ref(false);

/**
 * 导航菜单配置
 */
const navigation = [
  {
    name: "概览",
    path: "/",
    icon: LayoutDashboard,
  },
  {
    name: "供应商管理",
    path: "/providers",
    icon: Building2,
  },
  {
    name: "Agents 管理",
    path: "/agents",
    icon: Bot,
  },
  {
    name: "Tools 管理",
    path: "/tools",
    icon: Wrench,
  },
  {
    name: "MCP 管理",
    path: "/mcp",
    icon: Server,
  },
];

/**
 * 当前页面标题
 */
const pageTitle = computed(() => {
  const currentNav = navigation.find((item) => isActive(item.path));
  return currentNav?.name || "后台管理";
});

/**
 * 判断路由是否激活
 * @param {string} path - 路由路径
 * @returns {boolean} 是否激活
 */
function isActive(path) {
  if (path === "/") {
    return route.path === "/";
  }
  return route.path.startsWith(path);
}

/**
 * 切换主题
 */
function toggleTheme() {
  isDark.value = !isDark.value;
  if (isDark.value) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
}

// 监听路由变化，关闭移动端侧边栏
watch(
  () => route.path,
  () => {
    sidebarOpen.value = false;
  }
);
</script>
