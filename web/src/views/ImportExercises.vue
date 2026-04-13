<script setup lang="ts">
import { ref, computed, shallowRef, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { apiClient } from "@/lib/api";
import { exerciseKeys } from "@/lib/queryKeys";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { Textarea } from "@/components/ui/textarea";
import { useMuscleGroupOptions } from "@/features/reference/composables/useMuscleGroupOptions";
import { useEquipmentOptions } from "@/features/reference/composables/useEquipmentOptions";
import { useExerciseFeatureOptions } from "@/features/reference/composables/useExerciseFeatureOptions";
import type { FreeExercise, MappedImportExercise } from "@/features/exercises/import/types";
import { mapExercise } from "@/features/exercises/import/mapping";

const IMAGE_BASE = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/";
function thumbUrl(ex: FreeExercise): string | null {
  return ex.images.length > 0 ? `${IMAGE_BASE}${ex.images[0]}` : null;
}

const router = useRouter();
const queryClient = useQueryClient();

const { options: muscleGroupOptions } = useMuscleGroupOptions();
const { options: equipmentOptions } = useEquipmentOptions();
const { options: featureOptions } = useExerciseFeatureOptions();

const CATEGORY_OPTIONS_EDIT = [
  { value: "strength", label: "Strength" },
  { value: "cardio", label: "Cardio" },
  { value: "stretching", label: "Stretching" },
];
const FORCE_OPTIONS_EDIT = [
  { value: "pull", label: "Pull" },
  { value: "push", label: "Push" },
  { value: "static", label: "Static" },
];

type Step = "select" | "review";
const step = ref<Step>("select");

// --- Data loading ---
const allExercises = shallowRef<FreeExercise[]>([]);
const loadError = ref<string | null>(null);
const loading = ref(false);

onMounted(async () => {
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
});

// --- Filter state ---
const searchQuery = ref("");
const categoryFilter = ref<string[]>([]);
const equipmentFilter = ref<string[]>([]);
const muscleFilter = ref<string[]>([]);
const forceFilter = ref<string[]>([]);
const levelFilter = ref<string[]>([]);

const CATEGORY_OPTIONS = [
  "strength", "cardio", "stretching", "powerlifting",
  "olympic weightlifting", "strongman", "plyometrics",
].map((v) => ({ value: v, label: v.charAt(0).toUpperCase() + v.slice(1) }));

const EQUIPMENT_OPTIONS = [
  "bands", "barbell", "body only", "cable", "dumbbell", "e-z curl bar",
  "exercise ball", "foam roll", "kettlebells", "machine", "medicine ball", "other",
].map((v) => ({ value: v, label: v.charAt(0).toUpperCase() + v.slice(1) }));

const MUSCLE_OPTIONS = [
  "abdominals", "abductors", "adductors", "biceps", "calves", "chest",
  "forearms", "glutes", "hamstrings", "lats", "lower back", "middle back",
  "neck", "quadriceps", "shoulders", "traps", "triceps",
].map((v) => ({ value: v, label: v.charAt(0).toUpperCase() + v.slice(1) }));

const FORCE_OPTIONS = ["pull", "push", "static"]
  .map((v) => ({ value: v, label: v.charAt(0).toUpperCase() + v.slice(1) }));

const LEVEL_OPTIONS = ["beginner", "intermediate", "expert"]
  .map((v) => ({ value: v, label: v.charAt(0).toUpperCase() + v.slice(1) }));

const hasActiveFilters = computed(() =>
  searchQuery.value ||
  categoryFilter.value.length ||
  equipmentFilter.value.length ||
  muscleFilter.value.length ||
  forceFilter.value.length ||
  levelFilter.value.length
);

function clearFilters() {
  searchQuery.value = "";
  categoryFilter.value = [];
  equipmentFilter.value = [];
  muscleFilter.value = [];
  forceFilter.value = [];
  levelFilter.value = [];
}

// --- Filtered results ---
const filtered = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  const cats = categoryFilter.value;
  const equip = equipmentFilter.value;
  const muscles = muscleFilter.value;
  const forces = forceFilter.value;
  const levels = levelFilter.value;

  return allExercises.value.filter((ex) => {
    if (q && !ex.name.toLowerCase().includes(q)) return false;
    if (cats.length && !cats.includes(ex.category)) return false;
    if (equip.length && (!ex.equipment || !equip.includes(ex.equipment))) return false;
    if (muscles.length) {
      const exMuscles = [...ex.primaryMuscles, ...ex.secondaryMuscles];
      if (!muscles.some((m) => exMuscles.includes(m))) return false;
    }
    if (forces.length && (!ex.force || !forces.includes(ex.force))) return false;
    if (levels.length && !levels.includes(ex.level)) return false;
    return true;
  });
});

const VISIBLE_LIMIT = 100;
const visibleExercises = computed(() => filtered.value.slice(0, VISIBLE_LIMIT));

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

function toggleAllVisible() {
  const ids = visibleExercises.value.map((e) => e.id);
  const allSelected = ids.every((id) => selectedIds.value.has(id));
  const s = new Set(selectedIds.value);
  if (allSelected) ids.forEach((id) => s.delete(id));
  else ids.forEach((id) => s.add(id));
  selectedIds.value = s;
}

const visibleAllSelected = computed(() =>
  visibleExercises.value.length > 0 &&
  visibleExercises.value.every((e) => selectedIds.value.has(e.id))
);

const selectedCount = computed(() => selectedIds.value.size);

// --- Review ---
// Mutable snapshot taken when entering review step
const reviewList = ref<MappedImportExercise[]>([]);
const expandedIds = ref<Set<string>>(new Set());

function enterReview() {
  reviewList.value = allExercises.value
    .filter((ex) => selectedIds.value.has(ex.id))
    .map(mapExercise);
  expandedIds.value = new Set();
  step.value = "review";
}

function removeFromReview(id: string) {
  reviewList.value = reviewList.value.filter((ex) => ex.source.id !== id);
  const s = new Set(selectedIds.value);
  s.delete(id);
  selectedIds.value = s;
}

function toggleExpand(id: string) {
  const s = new Set(expandedIds.value);
  if (s.has(id)) s.delete(id);
  else s.add(id);
  expandedIds.value = s;
}

function isExpanded(id: string) {
  return expandedIds.value.has(id);
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

async function runImport() {
  importing.value = true;
  importProgress.value = 0;
  importErrors.value = [];
  const exercises = reviewList.value;
  importTotal.value = exercises.length;

  for (const ex of exercises) {
    try {
      let imageFile: File | null = null;
      if (ex.imageUrl) imageFile = await fetchImageAsFile(ex.imageUrl, ex.name);

      if (imageFile) {
        const fd = new FormData();
        fd.append("name", ex.name);
        if (ex.description) fd.append("description", ex.description);
        if (ex.force) fd.append("force", ex.force);
        if (ex.category) fd.append("category", ex.category);
        ex.instructions.forEach((i) => fd.append("instructions", i));
        ex.primary_muscle_groups.forEach((mg) => fd.append("primary_muscle_groups", mg));
        ex.secondary_muscle_groups.forEach((mg) => fd.append("secondary_muscle_groups", mg));
        ex.equipment.forEach((eq) => fd.append("equipment", eq));
        ex.exercise_features.forEach((ef) => fd.append("exercise_features", ef));
        fd.append("image", imageFile);
        await apiClient.post("/exercises", fd);
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
    } catch {
      importErrors.value.push(ex.name);
    }
    importProgress.value++;
  }

  importing.value = false;
  await queryClient.invalidateQueries({ queryKey: exerciseKeys.all });

  const succeeded = importTotal.value - importErrors.value.length;
  if (succeeded > 0) toast.success(`Imported ${succeeded} exercise${succeeded !== 1 ? "s" : ""}`);
  if (importErrors.value.length > 0) toast.error(`${importErrors.value.length} failed to import`);

  router.push({ name: "exercises" });
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <!-- Header -->
    <div class="flex items-start justify-between gap-4">
      <div>
        <button
          class="text-sm text-muted-foreground hover:text-foreground transition-colors mb-1"
          @click="step === 'review' ? step = 'select' : router.push({ name: 'exercises' })"
        >
          {{ step === "review" ? "← Back to selection" : "← Exercises" }}
        </button>
        <h1 class="text-3xl font-bold">
          {{ step === "select" ? "Import Exercises" : "Review Import" }}
        </h1>
        <p class="text-muted-foreground mt-1">
          <template v-if="step === 'select'">
            Browse and select exercises from the free exercise database to import.
          </template>
          <template v-else>
            Review {{ reviewList.length }} exercise{{ reviewList.length !== 1 ? "s" : "" }} before importing. Click any row to edit details.
          </template>
        </p>
      </div>
      <div v-if="step === 'select'" class="flex items-center gap-3 shrink-0 pt-1">
        <span class="text-sm text-muted-foreground hidden sm:block">{{ selectedCount }} selected</span>
        <Button :disabled="selectedCount === 0" @click="enterReview">
          Review<span class="hidden sm:inline">&nbsp;{{ selectedCount > 0 ? selectedCount : "" }} exercise{{ selectedCount !== 1 ? "s" : "" }}</span>
          <span class="sm:hidden">&nbsp;({{ selectedCount }})</span>
        </Button>
      </div>
      <div v-else class="flex items-center gap-3 shrink-0 pt-1">
        <Button variant="outline" :disabled="importing" @click="step = 'select'">Back</Button>
        <Button :disabled="reviewList.length === 0 || importing" @click="runImport">
          {{ importing ? `${importProgress}/${importTotal}` : `Import ${reviewList.length}` }}
        </Button>
      </div>
    </div>

    <!-- Import progress bar -->
    <div v-if="importing" class="rounded-lg border bg-muted/30 px-4 py-3">
      <div class="flex items-center justify-between text-sm mb-2">
        <span class="text-muted-foreground">Importing exercises…</span>
        <span class="tabular-nums text-muted-foreground">{{ importProgress }} / {{ importTotal }}</span>
      </div>
      <div class="h-2 bg-muted rounded-full overflow-hidden">
        <div
          class="h-full bg-primary rounded-full transition-all duration-300"
          :style="{ width: `${importTotal ? (importProgress / importTotal) * 100 : 0}%` }"
        />
      </div>
    </div>

    <!-- SELECT STEP -->
    <template v-if="step === 'select'">
      <div v-if="loading" class="py-24 text-center text-muted-foreground text-sm">
        Loading exercise database…
      </div>
      <div v-else-if="loadError" class="py-24 text-center text-destructive text-sm">
        {{ loadError }}
      </div>
      <template v-else>
        <!-- Filters -->
        <div class="flex flex-col gap-4 p-4 border rounded-lg bg-card">
          <div class="flex flex-col sm:flex-row gap-3">
            <Input
              v-model="searchQuery"
              placeholder="Search exercises…"
              class="sm:flex-1"
            />
            <MultiSelectTags
              v-model="categoryFilter"
              :options="CATEGORY_OPTIONS"
              placeholder="Category"
              class="sm:flex-1"
            />
            <MultiSelectTags
              v-model="equipmentFilter"
              :options="EQUIPMENT_OPTIONS"
              placeholder="Equipment"
              class="sm:flex-1"
            />
            <MultiSelectTags
              v-model="muscleFilter"
              :options="MUSCLE_OPTIONS"
              placeholder="Muscle group"
              class="sm:flex-1"
            />
            <MultiSelectTags
              v-model="forceFilter"
              :options="FORCE_OPTIONS"
              placeholder="Force"
              class="sm:flex-1"
            />
            <MultiSelectTags
              v-model="levelFilter"
              :options="LEVEL_OPTIONS"
              placeholder="Level"
              class="sm:flex-1"
            />
          </div>
          <div class="flex gap-2">
            <Button variant="outline" class="flex-1 sm:flex-none" :disabled="!hasActiveFilters" @click="clearFilters">
              Clear
            </Button>
          </div>
        </div>

        <!-- Exercise list -->
        <div class="rounded-lg border overflow-hidden">
          <!-- Table header -->
          <div class="flex items-center gap-3 px-4 py-2.5 bg-muted/40 border-b text-xs font-medium text-muted-foreground">
            <Checkbox :model-value="visibleAllSelected" @update:model-value="toggleAllVisible" />
            <span class="flex-1">
              Exercise
              <span class="font-normal ml-1 text-muted-foreground/70">
                ({{ visibleExercises.length
                }}<template v-if="filtered.length > VISIBLE_LIMIT"> of {{ filtered.length }}</template>)
              </span>
            </span>
            <span class="w-32 hidden sm:block">Muscles</span>
            <span class="w-24 hidden md:block">Equipment</span>
            <span class="w-20 text-right hidden sm:block">Category</span>
          </div>

          <!-- Rows -->
          <div
            v-for="ex in visibleExercises"
            :key="ex.id"
            class="flex items-center gap-3 px-4 py-3 border-b last:border-0 hover:bg-muted/20 cursor-pointer transition-colors"
            :class="{ 'bg-primary/5': isSelected(ex.id) }"
            @click="toggleSelect(ex.id)"
          >
            <Checkbox
              :model-value="isSelected(ex.id)"
              @update:model-value="() => toggleSelect(ex.id)"
              @click.stop
            />
            <div class="w-9 h-9 shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center">
              <img
                v-if="thumbUrl(ex)"
                :src="thumbUrl(ex)!"
                :alt="ex.name"
                class="w-full h-full object-cover"
                loading="lazy"
              />
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground/40">
                <rect width="18" height="18" x="3" y="3" rx="2"/><path d="m9 9 6 6m0-6-6 6"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ ex.name }}</p>
              <p class="text-xs text-muted-foreground capitalize sm:hidden">
                {{ ex.primaryMuscles.slice(0, 2).join(", ") }}
                <span v-if="ex.equipment"> · {{ ex.equipment }}</span>
              </p>
            </div>
            <span class="w-32 text-xs text-muted-foreground hidden sm:block truncate capitalize">
              {{ ex.primaryMuscles.slice(0, 2).join(", ") }}
            </span>
            <span class="w-24 text-xs text-muted-foreground hidden md:block truncate capitalize">
              {{ ex.equipment ?? "—" }}
            </span>
            <span class="w-20 text-right text-xs text-muted-foreground capitalize shrink-0 hidden sm:block">
              {{ ex.category }}
            </span>
          </div>

          <div v-if="filtered.length > VISIBLE_LIMIT" class="px-4 py-3 text-xs text-muted-foreground text-center border-t bg-muted/20">
            {{ filtered.length - VISIBLE_LIMIT }} more — refine filters to narrow results
          </div>
          <div v-if="filtered.length === 0" class="py-16 text-center text-sm text-muted-foreground">
            No exercises match your filters
          </div>
        </div>
      </template>
    </template>

    <!-- REVIEW STEP -->
    <template v-else>
      <div v-if="reviewList.length === 0" class="py-16 text-center text-sm text-muted-foreground">
        No exercises selected
      </div>
      <div v-else class="rounded-lg border overflow-hidden">
        <div
          v-for="ex in reviewList"
          :key="ex.source.id"
          class="border-b last:border-0"
        >
          <!-- Summary row — click to expand -->
          <div
            class="flex items-center gap-3 px-4 sm:px-6 py-3 hover:bg-muted/20 cursor-pointer transition-colors"
            :class="{ 'bg-muted/10': isExpanded(ex.source.id) }"
            @click="toggleExpand(ex.source.id)"
          >
            <div class="w-9 h-9 shrink-0 rounded overflow-hidden bg-muted flex items-center justify-center">
              <img
                v-if="ex.imageUrl"
                :src="ex.imageUrl"
                :alt="ex.name"
                class="w-full h-full object-cover"
                loading="lazy"
              />
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground/40">
                <rect width="18" height="18" x="3" y="3" rx="2"/><path d="m9 9 6 6m0-6-6 6"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ ex.name }}</p>
              <p class="text-xs text-muted-foreground truncate">
                <span v-if="ex.category" class="capitalize">{{ ex.category }}</span>
                <span v-if="ex.category && ex.equipment.length"> · </span>
                <span class="capitalize">{{ ex.equipment.join(", ") }}</span>
                <span v-if="ex.primary_muscle_groups.length"> · {{ ex.primary_muscle_groups.join(", ") }}</span>
              </p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <div class="flex gap-1">
                <span
                  v-for="feat in ex.exercise_features"
                  :key="feat"
                  class="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary"
                >
                  {{ feat }}
                </span>
              </div>
              <svg
                xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                class="text-muted-foreground transition-transform"
                :class="{ 'rotate-180': isExpanded(ex.source.id) }"
              >
                <path d="m6 9 6 6 6-6"/>
              </svg>
              <button
                class="text-muted-foreground hover:text-destructive transition-colors p-1"
                :disabled="importing"
                @click.stop="removeFromReview(ex.source.id)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Edit form — shown when expanded -->
          <div v-if="isExpanded(ex.source.id)" class="px-4 sm:px-6 pb-5 pt-1 border-t bg-muted/5">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-3">
              <!-- Name -->
              <div class="sm:col-span-2">
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Name</label>
                <Input v-model="ex.name" placeholder="Exercise name" />
              </div>

              <!-- Category -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Category</label>
                <select
                  v-model="ex.category"
                  class="w-full border border-input bg-background text-sm rounded-md px-3 py-2 focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  <option value="">None</option>
                  <option v-for="opt in CATEGORY_OPTIONS_EDIT" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- Force -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Force</label>
                <select
                  v-model="ex.force"
                  class="w-full border border-input bg-background text-sm rounded-md px-3 py-2 focus:outline-none focus:ring-1 focus:ring-ring"
                >
                  <option value="">None</option>
                  <option v-for="opt in FORCE_OPTIONS_EDIT" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
              </div>

              <!-- Primary muscles -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Primary muscles</label>
                <MultiSelectTags
                  v-model="ex.primary_muscle_groups"
                  :options="muscleGroupOptions"
                  placeholder="Select muscles…"
                />
              </div>

              <!-- Secondary muscles -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Secondary muscles</label>
                <MultiSelectTags
                  v-model="ex.secondary_muscle_groups"
                  :options="muscleGroupOptions"
                  placeholder="Select muscles…"
                />
              </div>

              <!-- Equipment -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Equipment</label>
                <MultiSelectTags
                  v-model="ex.equipment"
                  :options="equipmentOptions"
                  placeholder="Select equipment…"
                />
              </div>

              <!-- Features -->
              <div>
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Features</label>
                <MultiSelectTags
                  v-model="ex.exercise_features"
                  :options="featureOptions"
                  placeholder="Select features…"
                />
              </div>

              <!-- Instructions -->
              <div class="sm:col-span-2">
                <label class="text-xs font-medium text-muted-foreground mb-1.5 block">Instructions</label>
                <div class="space-y-2">
                  <div v-for="(_, i) in ex.instructions" :key="i" class="flex gap-2 items-start">
                    <span class="text-xs text-muted-foreground mt-2.5 w-5 shrink-0 text-right">{{ i + 1 }}.</span>
                    <Textarea
                      :model-value="ex.instructions[i]"
                      :placeholder="`Step ${i + 1}…`"
                      rows="2"
                      class="flex-1 resize-none"
                      @input="(e: Event) => ex.instructions[i] = (e.target as HTMLTextAreaElement).value"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      class="mt-0.5 shrink-0"
                      @click="ex.instructions.splice(i, 1)"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                    </Button>
                  </div>
                  <Button type="button" variant="outline" size="sm" @click="ex.instructions.push('')">
                    + Add step
                  </Button>
                </div>
              </div>
            </div>
            <div class="mt-3 flex justify-end">
              <Button size="sm" variant="outline" @click="toggleExpand(ex.source.id)">Done</Button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
