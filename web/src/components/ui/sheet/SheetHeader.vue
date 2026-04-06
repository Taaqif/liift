<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{ class?: HTMLAttributes["class"] }>()
</script>

<template>
  <div data-slot="sheet-header" :class="cn('flex flex-col gap-1.5 px-6 pt-6 pb-4 border-b shrink-0', props.class)">
    <slot />
  </div>
</template>
