import { createRouter, createWebHistory } from "vue-router";
import CommingSoon from "../views/ComingSoon.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "home",
      component: CommingSoon,
    },
    {
      path: "/workouts",
      name: "workouts",
      component: () => import("../views/Workouts.vue"),
    },
  ],
});

export default router;
