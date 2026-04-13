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
      path: "/workouts/active",
      name: "active-workout",
      component: () => import("../views/ActiveWorkout.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workouts/new",
      name: "workout-create",
      component: () => import("../views/WorkoutForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workouts/:id/edit",
      name: "workout-edit",
      component: () => import("../views/WorkoutForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/exercises",
      name: "exercises",
      component: () => import("../views/Exercises.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/exercises/import",
      name: "exercise-import",
      component: () => import("../views/ImportExercises.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/exercises/new",
      name: "exercise-create",
      component: () => import("../views/ExerciseForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/exercises/:id/edit",
      name: "exercise-edit",
      component: () => import("../views/ExerciseForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-plans",
      name: "workout-plans",
      component: () => import("../views/WorkoutPlans.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-plans/new",
      name: "workout-plan-create",
      component: () => import("../views/WorkoutPlanForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-plans/:id",
      name: "workout-plan-detail",
      component: () => import("../views/WorkoutPlanDetail.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-plans/:id/edit",
      name: "workout-plan-edit",
      component: () => import("../views/WorkoutPlanForm.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-plans/active",
      name: "active-plan",
      component: () => import("../views/ActivePlanView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/workout-history",
      name: "workout-history",
      component: () => import("../views/WorkoutHistory.vue"),
      meta: { requiresAuth: true },
    },
  ],
});

// Navigation guard
router.beforeEach((to, _from, next) => {
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
