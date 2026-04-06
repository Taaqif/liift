<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{ class?: HTMLAttributes["class"] }>()
</script>

<template>
  <div data-slot="sheet-footer" :class="cn('px-6 py-3 border-t shrink-0', props.class)">
    <slot />
  </div>
</template>
