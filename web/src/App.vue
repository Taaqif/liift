<script setup lang="ts">
import { onMounted } from "vue";
import { VueQueryDevtools } from "@tanstack/vue-query-devtools";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Toaster } from "@/components/ui/sonner";
import { useAuth } from "@/lib/auth/composables/useAuth";

const { user, isAuthenticated, logout, initAuth } = useAuth();

onMounted(() => {
  initAuth();
});
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <div class="px-8 py-4 shadow flex items-center justify-between">
      <NavigationMenu>
        <NavigationMenuList>
          <NavigationMenuItem>
            <NavigationMenuLink as-child>
              <router-link
                to="/"
                active-class="text-accent-foreground bg-accent"
              >
                {{ $t("nav.home") }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavigationMenuLink as-child>
              <router-link
                to="/workouts"
                active-class="text-accent-foreground bg-accent"
              >
                {{ $t("nav.workouts") }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavigationMenuLink as-child>
              <router-link
                to="/workouts/active"
                active-class="text-accent-foreground bg-accent"
              >
                {{ $t("workoutSession.activeWorkout") }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavigationMenuLink as-child>
              <router-link
                to="/workout-plans"
                active-class="text-accent-foreground bg-accent"
              >
                {{ $t("nav.workoutPlans") }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavigationMenuLink as-child>
              <router-link
                to="/exercises"
                active-class="text-accent-foreground bg-accent"
              >
                {{ $t("nav.exercises") }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
        </NavigationMenuList>
      </NavigationMenu>
      <div v-if="isAuthenticated && user" class="flex items-center gap-4">
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" class="gap-2">
              <span>{{ user.username }}</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>
              <div class="flex flex-col">
                <span>{{ user.username }}</span>
                <span v-if="user.email" class="text-xs text-muted-foreground">
                  {{ user.email }}
                </span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="logout">
              {{ $t("auth.logOut") }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div v-else class="flex items-center gap-2">
        <Button variant="ghost" as-child>
          <router-link to="/login">{{ $t("auth.login") }}</router-link>
        </Button>
        <Button as-child>
          <router-link to="/register">{{ $t("auth.register") }}</router-link>
        </Button>
      </div>
    </div>
    <main class="p-2 md:p-8 flex-1">
      <router-view />
    </main>
    <VueQueryDevtools />
    <Toaster />
  </div>
</template>
