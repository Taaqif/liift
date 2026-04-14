<script setup lang="ts">
import { CheckIcon, ChevronDown, X } from "lucide-vue-next";
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
import { Input } from "@/components/ui/input";
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
    listMode?: boolean;
    single?: boolean;
  }>(),
  {
    placeholder: "Select options...",
    disabled: false,
    listMode: false,
    single: false,
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

function toggleOption(value: string) {
  if (props.single) {
    // Radio behavior: selecting the same item again deselects it
    selectedValues.value = selectedValues.value.includes(value) ? [] : [value];
    return;
  }
  const current = [...selectedValues.value];
  const idx = current.indexOf(value);
  if (idx >= 0) current.splice(idx, 1);
  else current.push(value);
  selectedValues.value = current;
}
</script>

<template>
  <!-- Inline list mode (tablet+ sidebar) -->
  <div v-if="listMode" class="hidden md:flex flex-col gap-1.5">
    <div class="flex items-center justify-between min-h-[1.25rem]">
      <span class="text-xs font-medium text-muted-foreground uppercase tracking-wide">{{ placeholder }}</span>
      <button
        v-if="selectedValues.length > 0"
        type="button"
        class="text-xs text-muted-foreground hover:text-foreground flex items-center gap-0.5"
        @click="selectedValues = []"
      >
        <X class="h-3 w-3" />{{ selectedValues.length }}
      </button>
    </div>
    <Input v-model="searchTerm" class="h-7 text-xs" placeholder="Filter..." :disabled="disabled" />
    <div class="max-h-44 overflow-y-auto space-y-0.5">
      <button
        v-for="item in filteredOptions"
        :key="item.value"
        type="button"
        :disabled="disabled"
        class="w-full flex items-center gap-2 px-2 py-1.5 text-sm rounded-sm hover:bg-accent hover:text-accent-foreground text-left disabled:pointer-events-none disabled:opacity-50"
        @click="toggleOption(item.value)"
      >
        <!-- Radio indicator -->
        <template v-if="single">
          <div
            :class="[
              'h-3.5 w-3.5 rounded-full border flex items-center justify-center shrink-0 transition-colors',
              selectedValues.includes(item.value)
                ? 'bg-primary border-primary text-primary-foreground'
                : 'border-input',
            ]"
          >
            <div v-if="selectedValues.includes(item.value)" class="h-1.5 w-1.5 rounded-full bg-primary-foreground" />
          </div>
        </template>
        <!-- Checkbox indicator -->
        <template v-else>
          <div
            :class="[
              'h-3.5 w-3.5 rounded-[3px] border flex items-center justify-center shrink-0 transition-colors',
              selectedValues.includes(item.value)
                ? 'bg-primary border-primary text-primary-foreground'
                : 'border-input',
            ]"
          >
            <CheckIcon v-if="selectedValues.includes(item.value)" class="h-2.5 w-2.5" />
          </div>
        </template>
        <span class="truncate">{{ item.label }}</span>
      </button>
    </div>
  </div>

  <!-- Popover mode: full-width on mobile, hidden on tablet+ when listMode -->
  <div :class="listMode ? 'md:hidden' : ''">
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
              :value="item.value" @select="() => { searchTerm = ''; }">
              <span>{{ item.label }}</span>

              <ListboxItemIndicator class="ml-auto inline-flex items-center justify-center">
                <CheckIcon />
              </ListboxItemIndicator>
            </ListboxItem>
          </ListboxContent>
        </PopoverContent>
      </ListboxRoot>
    </Popover>
  </div>
</template>
