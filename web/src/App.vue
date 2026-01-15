<script setup lang="ts">
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu";
</script>

<template>
  <div class="px-8 py-4 shadow">
    <NavigationMenu>
      <NavigationMenuList>
        <NavigationMenuItem>
          <NavigationMenuLink as-child>
            <router-link to="/" active-class="text-accent-foreground bg-accent">
              Home
            </router-link>
          </NavigationMenuLink>
        </NavigationMenuItem>
        <NavigationMenuItem>
          <NavigationMenuLink as-child>
            <router-link
              to="/workouts"
              active-class="text-accent-foreground bg-accent"
            >
              Workouts
            </router-link>
          </NavigationMenuLink>
        </NavigationMenuItem>
      </NavigationMenuList>
    </NavigationMenu>
  </div>
  <main class="p-8">
    <router-view />
  </main>
</template>
