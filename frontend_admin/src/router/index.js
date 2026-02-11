import { createRouter, createWebHistory } from "vue-router";
import AdminLayout from "@/layouts/AdminLayout.vue";

/**
 * 路由配置
 */
const routes = [
  {
    path: "/",
    component: AdminLayout,
    children: [
      {
        path: "",
        name: "Dashboard",
        component: () => import("@/views/dashboard/index.vue"),
      },
      {
        path: "providers",
        name: "Providers",
        component: () => import("@/views/providers/index.vue"),
      },
      {
        path: "agents",
        name: "Agents",
        component: () => import("@/views/agents/index.vue"),
      },
      {
        path: "tools",
        name: "Tools",
        component: () => import("@/views/tools/index.vue"),
      },
    ],
  },
];

/**
 * 创建路由实例
 */
const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
