<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { Button } from "@/components/ui/button";

interface Props {
  total: number;
  limit: number;
  offset: number;
  itemLabel?: string;
  itemLabelKey?: string; // i18n key for pluralization (e.g., "exercises.titleLower")
}

const props = withDefaults(defineProps<Props>(), {
  itemLabel: undefined,
  itemLabelKey: undefined,
});

const emit = defineEmits<{
  (e: "update:offset", value: number): void;
  (e: "page-change", page: number): void;
}>();

const { t } = useI18n();

const totalPages = computed(() => Math.ceil(props.total / props.limit));
const currentPage = computed(() => Math.floor(props.offset / props.limit) + 1);
const from = computed(() => props.offset + 1);
const to = computed(() => Math.min(props.offset + props.limit, props.total));
const hasNextPage = computed(() => props.offset + props.limit < props.total);
const hasPrevPage = computed(() => props.offset > 0);

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  const newOffset = (page - 1) * props.limit;
  emit("update:offset", newOffset);
  emit("page-change", page);
};

const itemCountLabel = computed(() => {
  const count = props.total;

  if (props.itemLabelKey) {
    const label = t(props.itemLabelKey, count);
    return label;
  }

  if (props.itemLabel) {
    return props.itemLabel;
  }

  const label = t("pagination.itemCount", count);
  return label;
});
</script>

<template>
  <div v-if="total > 0" class="flex items-center justify-between">
    <div class="text-sm text-muted-foreground">
      <span>
        {{
          t("pagination.showingFromToOfTotal", {
            from,
            to,
            total,
          })
        }}
      </span>
      <span class="ml-1">
        {{ itemCountLabel }}
      </span>
    </div>
    <div class="flex gap-2">
      <Button
        variant="outline"
        size="sm"
        @click="goToPage(currentPage - 1)"
        :disabled="!hasPrevPage"
      >
        {{ t("pagination.previous") }}
      </Button>
      <Button
        variant="outline"
        size="sm"
        @click="goToPage(currentPage + 1)"
        :disabled="!hasNextPage"
      >
        {{ t("pagination.next") }}
      </Button>
    </div>
  </div>
</template>
