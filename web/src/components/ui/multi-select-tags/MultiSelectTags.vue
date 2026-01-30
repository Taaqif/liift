<script setup lang="ts">
import { CheckIcon, ChevronDown } from "lucide-vue-next";
import {
  ListboxContent,
  ListboxFilter,
  ListboxItem,
  ListboxItemIndicator,
  ListboxRoot,
  useFilter,
  type AcceptableInputValue,
  type AcceptableValue,
} from "reka-ui";
import { computed, ref, watch } from "vue";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  Popover,
  PopoverAnchor,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  TagsInput,
  TagsInputInput,
  TagsInputItem,
  TagsInputItemDelete,
} from "@/components/ui/tags-input";

export type MultiSelectOption = {
  value: string;
  label: string;
};

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    options: MultiSelectOption[] | string[];
    placeholder?: string;
    class?: string;
    disabled?: boolean;
  }>(),
  {
    placeholder: "Select options...",
    disabled: false,
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
}>();

const searchTerm = ref("");
const open = ref(false);

const { contains } = useFilter({ sensitivity: "base" });

// Normalize options to always have value and label
const normalizedOptions = computed<MultiSelectOption[]>(() => {
  return props.options.map((option) => {
    if (typeof option === "string") {
      return { value: option, label: option };
    }
    return option;
  });
});

const filteredOptions = computed(() =>
  searchTerm.value === ""
    ? normalizedOptions.value
    : normalizedOptions.value.filter((option) =>
      contains(option.label, searchTerm.value),
    ),
);

watch(searchTerm, (f) => {
  if (f) {
    open.value = true;
  }
});

const selectedValues = computed({
  get: () => props.modelValue,
  set: (value) => emits("update:modelValue", value),
});

const getLabel = (value: string) =>
  normalizedOptions.value.find((o) => o.value === value)?.label ?? value;
</script>

<template>
  <Popover v-model:open="open">
    <ListboxRoot v-model="selectedValues as AcceptableValue" highlight-on-hover multiple :disabled="disabled">
      <PopoverAnchor class="inline-flex w-full">
        <PopoverTrigger as-child>
          <TagsInput v-slot="{ modelValue: tags }" v-model="selectedValues as AcceptableInputValue[]"
            :class="cn('w-full', props.class)" :disabled="disabled">
            <TagsInputItem v-for="item in tags" :key="item.toString()" :value="item.toString()">
              <span class="py-0.5 px-2 text-sm rounded bg-transparent">{{ getLabel(item.toString()) }}</span>
              <TagsInputItemDelete @click.stop />
            </TagsInputItem>

            <ListboxFilter v-model="searchTerm" as-child>
              <TagsInputInput :placeholder="placeholder" @keydown.enter.prevent @keydown.down="open = true" />
            </ListboxFilter>

            <Button size="icon-sm" variant="ghost" class="order-last self-start ml-auto" :disabled="disabled">
              <ChevronDown class="size-3.5" />
            </Button>
          </TagsInput>
        </PopoverTrigger>
      </PopoverAnchor>

      <PopoverContent class="p-1" @open-auto-focus.prevent>
        <ListboxContent
          class="max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto empty:after:content-['No_options'] empty:p-1 empty:after:block"
          tabindex="0">
          <ListboxItem v-for="item in filteredOptions" :key="item.value"
            class="data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground [&_svg:not([class*='text-'])]:text-muted-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"
            :value="item.value" @select="
              () => {
                searchTerm = '';
              }
            ">
            <span>{{ item.label }}</span>

            <ListboxItemIndicator class="ml-auto inline-flex items-center justify-center">
              <CheckIcon />
            </ListboxItemIndicator>
          </ListboxItem>
        </ListboxContent>
      </PopoverContent>
    </ListboxRoot>
  </Popover>
</template>
