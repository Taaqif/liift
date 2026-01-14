<script setup lang="ts"></script>

<template>
  <nav>
    <router-link to="/">Home</router-link>
    <router-link to="/about">About</router-link>
  </nav>
  <main>
    <router-view />
  </main>
</template>

<style scoped>
nav {
  padding: 1rem 2rem;
  background-color: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  gap: 1rem;
}

nav a {
  color: #374151;
  text-decoration: none;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  transition: background-color 0.2s;
}

nav a:hover {
  background-color: #e5e7eb;
}

nav a.router-link-active {
  color: #3b82f6;
  background-color: #dbeafe;
}

main {
  padding: 2rem;
}
</style>
