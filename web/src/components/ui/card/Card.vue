<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import type { PrimitiveProps } from "reka-ui";
import { Primitive } from "reka-ui";
import { cn } from "@/lib/utils";

interface Props extends PrimitiveProps {
  class?: HTMLAttributes["class"];
}
const props = withDefaults(defineProps<Props>(), {
  as: "div",
});
</script>
s

<template>
  <Primitive
    data-slot="card"
    :as="props.as"
    :as-child="props.asChild"
    :class="
      cn(
        'bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm',
        props.class,
      )
    "
  >
    <slot />
  </Primitive>
</template>
