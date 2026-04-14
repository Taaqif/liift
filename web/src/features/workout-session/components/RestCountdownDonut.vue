<script setup lang="ts">
const props = defineProps<{
  remaining: number;
  total: number;
}>();

const CIRC = 2 * Math.PI * 8;
</script>

<template>
  <svg viewBox="0 0 20 20" class="size-5 scale-x-[-1]" aria-hidden="true">
    <circle
      cx="10"
      cy="10"
      r="8"
      fill="none"
      stroke-width="3"
      class="text-muted-foreground/20"
      stroke="currentColor"
    />
    <circle
      cx="10"
      cy="10"
      r="8"
      fill="none"
      stroke-width="3"
      stroke-linecap="round"
      class="text-green-500"
      stroke="currentColor"
      transform="rotate(-90, 10, 10)"
      :stroke-dasharray="CIRC"
      :stroke-dashoffset="CIRC * (1 - remaining / total)"
      style="transition: stroke-dashoffset 0.95s linear"
    />
  </svg>
</template>
