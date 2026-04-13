<script setup lang="ts">
import { ref, computed, shallowRef } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { apiClient } from "@/lib/api";
import { exerciseKeys } from "@/lib/queryKeys";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import type { FreeExercise, MappedImportExercise } from "./types";
import { mapExercise } from "./mapping";

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ "update:open": [boolean] }>();

const queryClient = useQueryClient();

type Step = "search" | "review";
const step = ref<Step>("search");

// --- Data loading ---
const allExercises = shallowRef<FreeExercise[]>([]);
const loadError = ref<string | null>(null);
const loading = ref(false);

async function loadExercises() {
  if (allExercises.value.length > 0) return;
  loading.value = true;
  loadError.value = null;
  try {
    const res = await fetch("/exercises.json");
    if (!res.ok) throw new Error("Failed to load exercise database");
    allExercises.value = await res.json();
  } catch (e) {
    loadError.value = e instanceof Error ? e.message : "Unknown error";
  } finally {
    loading.value = false;
  }
}

// Load when dialog opens
import { watch } from "vue";
watch(
  () => props.open,
  (open) => {
    if (open) {
      loadExercises();
      step.value = "search";
    }
  }
);

// --- Search & filter ---
const searchQuery = ref("");
const categoryFilter = ref<string>("");

const categories = ["strength", "cardio", "stretching", "powerlifting", "olympic weightlifting", "strongman", "plyometrics"];

const filtered = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  const cat = categoryFilter.value;
  return allExercises.value.filter((ex) => {
    if (q && !ex.name.toLowerCase().includes(q)) return false;
    if (cat && ex.category !== cat) return false;
    return true;
  });
});

// --- Selection ---
const selectedIds = ref<Set<string>>(new Set());

function toggleSelect(id: string) {
  const s = new Set(selectedIds.value);
  if (s.has(id)) s.delete(id);
  else s.add(id);
  selectedIds.value = s;
}

function isSelected(id: string) {
  return selectedIds.value.has(id);
}

function toggleAll() {
  const visibleIds = filtered.value.slice(0, 50).map((e) => e.id);
  const allSelected = visibleIds.every((id) => selectedIds.value.has(id));
  const s = new Set(selectedIds.value);
  if (allSelected) {
    visibleIds.forEach((id) => s.delete(id));
  } else {
    visibleIds.forEach((id) => s.add(id));
  }
  selectedIds.value = s;
}

const selectedCount = computed(() => selectedIds.value.size);

// --- Review step ---
const mappedExercises = computed<MappedImportExercise[]>(() => {
  const selectedSet = selectedIds.value;
  return allExercises.value
    .filter((ex) => selectedSet.has(ex.id))
    .map(mapExercise);
});

function removeFromReview(id: string) {
  const s = new Set(selectedIds.value);
  s.delete(id);
  selectedIds.value = s;
}

function goToReview() {
  if (selectedCount.value === 0) return;
  step.value = "review";
}

function goBack() {
  step.value = "search";
}

// --- Import ---
const importing = ref(false);
const importProgress = ref(0);
const importTotal = ref(0);
const importErrors = ref<string[]>([]);

async function fetchImageAsFile(url: string, name: string): Promise<File | null> {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    const blob = await res.blob();
    const ext = url.split(".").pop() ?? "jpg";
    return new File([blob], `${name}.${ext}`, { type: blob.type });
  } catch {
    return null;
  }
}

async function importExercises() {
  importing.value = true;
  importProgress.value = 0;
  importErrors.value = [];
  const exercises = mappedExercises.value;
  importTotal.value = exercises.length;

  for (const ex of exercises) {
    try {
      let imageFile: File | null = null;
      if (ex.imageUrl) {
        imageFile = await fetchImageAsFile(ex.imageUrl, ex.name);
      }

      if (imageFile) {
        const formData = new FormData();
        formData.append("name", ex.name);
        if (ex.description) formData.append("description", ex.description);
        if (ex.force) formData.append("force", ex.force);
        if (ex.category) formData.append("category", ex.category);
        ex.instructions.forEach((i) => formData.append("instructions", i));
        ex.primary_muscle_groups.forEach((mg) =>
          formData.append("primary_muscle_groups", mg)
        );
        ex.secondary_muscle_groups.forEach((mg) =>
          formData.append("secondary_muscle_groups", mg)
        );
        ex.equipment.forEach((eq) => formData.append("equipment", eq));
        ex.exercise_features.forEach((ef) =>
          formData.append("exercise_features", ef)
        );
        formData.append("image", imageFile);
        await apiClient.post("/exercises", formData);
      } else {
        await apiClient.post("/exercises", {
          name: ex.name,
          description: ex.description || undefined,
          force: ex.force,
          category: ex.category,
          instructions: ex.instructions,
          primary_muscle_groups: ex.primary_muscle_groups,
          secondary_muscle_groups: ex.secondary_muscle_groups,
          equipment: ex.equipment,
          exercise_features: ex.exercise_features,
        });
      }
    } catch (e) {
      importErrors.value.push(ex.name);
    }
    importProgress.value++;
  }

  importing.value = false;
  await queryClient.invalidateQueries({ queryKey: exerciseKeys.all });

  const succeeded = importTotal.value - importErrors.value.length;
  if (succeeded > 0) {
    toast.success(`Imported ${succeeded} exercise${succeeded !== 1 ? "s" : ""}`);
  }
  if (importErrors.value.length > 0) {
    toast.error(`${importErrors.value.length} exercise${importErrors.value.length !== 1 ? "s" : ""} failed to import`);
  }

  emit("update:open", false);
  selectedIds.value = new Set();
  step.value = "search";
}

const VISIBLE_LIMIT = 50;
const visibleExercises = computed(() => filtered.value.slice(0, VISIBLE_LIMIT));
const visibleAllSelected = computed(() =>
  visibleExercises.value.length > 0 &&
  visibleExercises.value.every((e) => selectedIds.value.has(e.id))
);
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-2xl h-[85dvh] flex flex-col gap-0 p-0 overflow-hidden">
      <DialogHeader class="shrink-0 px-6 pt-6 pb-4 border-b">
        <DialogTitle>
          {{ step === "search" ? "Import Exercises" : "Review Import" }}
        </DialogTitle>
        <DialogDescription>
          <span v-if="step === 'search'">
            Search the free exercise database and select exercises to import.
          </span>
          <span v-else>
            Review {{ mappedExercises.length }} exercise{{ mappedExercises.length !== 1 ? "s" : "" }} before importing.
          </span>
        </DialogDescription>
      </DialogHeader>

      <!-- SEARCH STEP -->
      <template v-if="step === 'search'">
        <div class="shrink-0 px-6 py-3 border-b flex gap-2">
          <Input
            v-model="searchQuery"
            placeholder="Search exercises..."
            class="flex-1"
          />
          <select
            v-model="categoryFilter"
            class="border border-input bg-background text-sm rounded-md px-3 py-2 focus:outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="">All categories</option>
            <option v-for="cat in categories" :key="cat" :value="cat">
              {{ cat.charAt(0).toUpperCase() + cat.slice(1) }}
            </option>
          </select>
        </div>

        <div v-if="loading" class="flex-1 flex items-center justify-center text-muted-foreground text-sm">
          Loading exercise database...
        </div>
        <div v-else-if="loadError" class="flex-1 flex items-center justify-center text-destructive text-sm px-6">
          {{ loadError }}
        </div>
        <div v-else class="flex-1 overflow-y-auto">
          <!-- Header row -->
          <div class="flex items-center gap-3 px-4 py-2 border-b bg-muted/40 text-xs text-muted-foreground font-medium sticky top-0">
            <Checkbox
              :model-value="visibleAllSelected"
              @update:model-value="toggleAll"
            />
            <span class="flex-1">Exercise (showing {{ visibleExercises.length }} of {{ filtered.length }})</span>
            <span class="w-24 text-right">Category</span>
          </div>

          <div
            v-for="ex in visibleExercises"
            :key="ex.id"
            class="flex items-center gap-3 px-4 py-3 border-b hover:bg-muted/30 cursor-pointer"
            @click="toggleSelect(ex.id)"
          >
            <Checkbox :model-value="isSelected(ex.id)" @update:model-value="() => toggleSelect(ex.id)" @click.stop />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium leading-tight">{{ ex.name }}</p>
              <p class="text-xs text-muted-foreground">
                {{ ex.primaryMuscles.join(", ") }}
                <span v-if="ex.equipment"> · {{ ex.equipment }}</span>
              </p>
            </div>
            <span class="w-24 text-right text-xs text-muted-foreground capitalize shrink-0">{{ ex.category }}</span>
          </div>

          <div v-if="filtered.length > VISIBLE_LIMIT" class="px-4 py-2 text-xs text-muted-foreground text-center border-b">
            {{ filtered.length - VISIBLE_LIMIT }} more results — refine your search to see them
          </div>
          <div v-if="filtered.length === 0" class="py-12 text-center text-sm text-muted-foreground">
            No exercises found
          </div>
        </div>

        <div class="shrink-0 px-6 py-4 border-t flex items-center justify-between">
          <span class="text-sm text-muted-foreground">
            {{ selectedCount }} selected
          </span>
          <Button :disabled="selectedCount === 0" @click="goToReview">
            Review {{ selectedCount > 0 ? selectedCount : "" }} exercise{{ selectedCount !== 1 ? "s" : "" }}
          </Button>
        </div>
      </template>

      <!-- REVIEW STEP -->
      <template v-else>
        <div class="flex-1 overflow-y-auto">
          <div
            v-for="ex in mappedExercises"
            :key="ex.source.id"
            class="border-b px-6 py-4"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1 min-w-0">
                <p class="font-medium text-sm">{{ ex.name }}</p>
                <div class="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                  <span v-if="ex.category">
                    <span class="font-medium text-foreground/60">Category:</span> {{ ex.category }}
                  </span>
                  <span v-if="ex.force">
                    <span class="font-medium text-foreground/60">Force:</span> {{ ex.force }}
                  </span>
                  <span>
                    <span class="font-medium text-foreground/60">Equipment:</span> {{ ex.equipment.join(", ") }}
                  </span>
                </div>
                <div class="mt-1 text-xs text-muted-foreground">
                  <span class="font-medium text-foreground/60">Primary:</span>
                  {{ ex.primary_muscle_groups.join(", ") }}
                </div>
                <div v-if="ex.secondary_muscle_groups.length > 0" class="mt-0.5 text-xs text-muted-foreground">
                  <span class="font-medium text-foreground/60">Secondary:</span>
                  {{ ex.secondary_muscle_groups.join(", ") }}
                </div>
                <div class="mt-1 flex gap-1 flex-wrap">
                  <span
                    v-for="feat in ex.exercise_features"
                    :key="feat"
                    class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary"
                  >
                    {{ feat }}
                  </span>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                class="shrink-0 text-muted-foreground hover:text-destructive"
                @click="removeFromReview(ex.source.id)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
                </svg>
              </Button>
            </div>
          </div>

          <div v-if="mappedExercises.length === 0" class="py-12 text-center text-sm text-muted-foreground">
            No exercises selected
          </div>
        </div>

        <!-- Import progress -->
        <div v-if="importing" class="shrink-0 px-6 py-3 border-t bg-muted/30">
          <div class="flex items-center justify-between text-sm mb-1">
            <span class="text-muted-foreground">Importing...</span>
            <span class="text-muted-foreground">{{ importProgress }} / {{ importTotal }}</span>
          </div>
          <div class="h-1.5 bg-muted rounded-full overflow-hidden">
            <div
              class="h-full bg-primary rounded-full transition-all duration-300"
              :style="{ width: `${(importProgress / importTotal) * 100}%` }"
            />
          </div>
        </div>

        <div class="shrink-0 px-6 py-4 border-t flex items-center justify-between">
          <Button variant="outline" :disabled="importing" @click="goBack">Back</Button>
          <Button
            :disabled="mappedExercises.length === 0 || importing"
            @click="importExercises"
          >
            {{ importing ? "Importing..." : `Import ${mappedExercises.length} exercise${mappedExercises.length !== 1 ? "s" : ""}` }}
          </Button>
        </div>
      </template>
    </DialogContent>
  </Dialog>
</template>
