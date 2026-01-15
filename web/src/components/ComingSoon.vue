<script setup lang="ts">
import { HammerIcon } from "lucide-vue-next";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
</script>

<template>
  <Empty>
    <EmptyHeader>
      <EmptyMedia variant="icon" class="size-20">
        <HammerIcon class="size-12" />
      </EmptyMedia>
    </EmptyHeader>
    <EmptyTitle>{{ $t("commingSoon") }}</EmptyTitle>
    <EmptyDescription>{{ $t("featureUnderConstruction") }}</EmptyDescription>
  </Empty>
</template>
