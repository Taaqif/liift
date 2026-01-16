import { createRouter, createWebHistory } from "vue-router";
import CommingSoon from "../views/ComingSoon.vue";
import { useAuth } from "@/lib/auth/composables/useAuth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("../views/Login.vue"),
      meta: { requiresGuest: true },
    },
    {
      path: "/register",
      name: "register",
      component: () => import("../views/Register.vue"),
      meta: { requiresGuest: true },
    },
    {
      path: "/",
      name: "home",
      component: CommingSoon,
      meta: { requiresAuth: true },
    },
    {
      path: "/workouts",
      name: "workouts",
      component: () => import("../views/Workouts.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/exercises",
      name: "exercises",
      component: () => import("../views/Exercises.vue"),
      meta: { requiresAuth: true },
    },
  ],
});

// Navigation guard
router.beforeEach((to, from, next) => {
  const { isAuthenticated } = useAuth();

  // Check if route requires authentication
  if (to.meta.requiresAuth && !isAuthenticated.value) {
    next({ name: "login", query: { redirect: to.fullPath } });
    return;
  }

  // Check if route requires guest (not authenticated)
  if (to.meta.requiresGuest && isAuthenticated.value) {
    next({ name: "home" });
    return;
  }

  next();
});

export default router;
