<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { useProfile } from "@/features/profile/composables/useProfile";
import { useUpdateProfile } from "@/features/profile/composables/useUpdateProfile";
import { GENDER_OPTIONS } from "@/features/profile/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const router = useRouter();
const { profile } = useProfile();
const { updateProfile, updating } = useUpdateProfile();

const step = ref(1);
const TOTAL_STEPS = 3;

// Step 1
const name = ref("");
// Step 2
const dateOfBirth = ref("");
const gender = ref("");
// Step 3
const weightKg = ref<string>("");
const heightCm = ref<string>("");

onMounted(() => {
  // Redirect if already onboarded
  if (profile.value?.onboarding_complete) {
    router.replace("/");
  }
});

const nameError = ref("");

function validateStep1() {
  if (!name.value.trim()) {
    nameError.value = "Name is required";
    return false;
  }
  nameError.value = "";
  return true;
}

async function goNext() {
  if (step.value === 1) {
    if (!validateStep1()) return;
  }
  if (step.value < TOTAL_STEPS) {
    step.value++;
  }
}

function goBack() {
  if (step.value > 1) step.value--;
}

async function finish() {
  if (!validateStep1()) {
    step.value = 1;
    return;
  }

  const payload: Record<string, unknown> = {
    name: name.value.trim(),
    onboarding_complete: true,
  };
  if (dateOfBirth.value) payload.date_of_birth = dateOfBirth.value;
  if (gender.value) payload.gender = gender.value;
  if (weightKg.value) payload.weight_kg = parseFloat(weightKg.value);
  if (heightCm.value) payload.height_cm = parseFloat(heightCm.value);

  try {
    await updateProfile(payload as any);
    router.push("/");
  } catch {
    toast.error("Failed to save profile. Please try again.");
  }
}

async function skip() {
  try {
    await updateProfile({ onboarding_complete: true });
    router.push("/");
  } catch {
    router.push("/");
  }
}

const progressPercent = computed(() => ((step.value - 1) / TOTAL_STEPS) * 100);
const canSkipAll = computed(() => step.value > 1);
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center">
    <div class="w-full max-w-md">

      <!-- Progress bar -->
      <div class="mb-8">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs text-muted-foreground">Step {{ step }} of {{ TOTAL_STEPS }}</span>
          <button
            v-if="canSkipAll"
            class="text-xs text-muted-foreground hover:text-foreground transition-colors"
            @click="skip"
          >
            Skip setup
          </button>
        </div>
        <div class="h-1 bg-muted rounded-full overflow-hidden">
          <div
            class="h-full bg-primary rounded-full transition-all duration-500"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
      </div>

      <!-- Step 1: Name -->
      <div v-if="step === 1" class="space-y-6">
        <div>
          <h1 class="text-3xl font-bold tracking-tight">What's your name?</h1>
          <p class="text-muted-foreground mt-1">This is how we'll greet you in the app.</p>
        </div>

        <div class="space-y-2">
          <Label for="name">Your name <span class="text-destructive">*</span></Label>
          <Input
            id="name"
            v-model="name"
            placeholder="e.g. Alex"
            class="text-base h-11"
            autofocus
            @keyup.enter="goNext"
          />
          <p v-if="nameError" class="text-sm text-destructive">{{ nameError }}</p>
        </div>

        <div class="flex gap-3">
          <Button class="flex-1 h-11" @click="goNext">
            Continue
          </Button>
        </div>
      </div>

      <!-- Step 2: Date of birth + gender -->
      <div v-else-if="step === 2" class="space-y-6">
        <div>
          <h1 class="text-3xl font-bold tracking-tight">A bit about you</h1>
          <p class="text-muted-foreground mt-1">Helps personalise your experience. All optional.</p>
        </div>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="dob">Date of birth</Label>
            <Input
              id="dob"
              v-model="dateOfBirth"
              type="date"
              class="h-11"
            />
          </div>

          <div class="space-y-2">
            <Label>Gender</Label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="opt in GENDER_OPTIONS"
                :key="opt.value"
                type="button"
                class="h-11 rounded-lg border text-sm font-medium transition-all"
                :class="gender === opt.value
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border bg-background text-muted-foreground hover:border-foreground/30 hover:text-foreground'"
                @click="gender = gender === opt.value ? '' : opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="flex gap-3">
          <Button variant="outline" class="h-11 px-5" @click="goBack">Back</Button>
          <Button class="flex-1 h-11" @click="goNext">Continue</Button>
        </div>
      </div>

      <!-- Step 3: Weight + height -->
      <div v-else-if="step === 3" class="space-y-6">
        <div>
          <h1 class="text-3xl font-bold tracking-tight">Your measurements</h1>
          <p class="text-muted-foreground mt-1">Used for tracking progress over time. Optional.</p>
        </div>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="weight">Weight (kg)</Label>
            <Input
              id="weight"
              v-model="weightKg"
              type="number"
              min="20"
              max="300"
              step="0.1"
              placeholder="e.g. 75"
              class="h-11"
            />
          </div>

          <div class="space-y-2">
            <Label for="height">Height (cm)</Label>
            <Input
              id="height"
              v-model="heightCm"
              type="number"
              min="100"
              max="250"
              step="0.5"
              placeholder="e.g. 178"
              class="h-11"
            />
          </div>
        </div>

        <div class="flex gap-3">
          <Button variant="outline" class="h-11 px-5" @click="goBack">Back</Button>
          <Button class="flex-1 h-11" :disabled="updating" @click="finish">
            {{ updating ? "Saving..." : "Finish setup" }}
          </Button>
        </div>
      </div>

    </div>
  </div>
</template>
