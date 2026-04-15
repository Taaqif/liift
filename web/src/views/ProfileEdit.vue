<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { useProfile } from "@/features/profile/composables/useProfile";
import { useUpdateProfile } from "@/features/profile/composables/useUpdateProfile";
import { GENDER_OPTIONS } from "@/features/profile/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const router = useRouter();
const { profile, loading } = useProfile();
const { updateProfile, updating } = useUpdateProfile();

const name = ref("");
const dateOfBirth = ref("");
const gender = ref("");
const weightKg = ref("");
const heightCm = ref("");

watch(
  profile,
  (p) => {
    if (!p) return;
    name.value = p.name ?? "";
    dateOfBirth.value = p.date_of_birth ?? "";
    gender.value = p.gender ?? "";
    weightKg.value = p.weight_kg != null ? String(p.weight_kg) : "";
    heightCm.value = p.height_cm != null ? String(p.height_cm) : "";
  },
  { immediate: true },
);

const nameError = ref("");

async function onSubmit() {
  if (!name.value.trim()) {
    nameError.value = "Name is required";
    return;
  }
  nameError.value = "";

  const payload: Record<string, unknown> = {
    name: name.value.trim(),
    date_of_birth: dateOfBirth.value || "",
    gender: gender.value || "",
    weight_kg: weightKg.value ? parseFloat(weightKg.value) : null,
    height_cm: heightCm.value ? parseFloat(heightCm.value) : null,
  };

  try {
    await updateProfile(payload as any);
    toast.success("Profile updated");
    router.push({ name: "profile" });
  } catch {
    toast.error("Failed to update profile");
  }
}
</script>

<template>
  <div class="max-w-lg">
    <div class="mb-8">
      <button
        class="text-sm text-muted-foreground hover:text-foreground transition-colors mb-1"
        @click="router.push({ name: 'profile' })"
      >
        ← Profile
      </button>
      <h1 class="text-3xl font-bold tracking-tight">Edit Profile</h1>
      <p class="text-muted-foreground mt-1">Update your personal details.</p>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground text-sm">Loading...</div>
    </div>

    <form v-else class="space-y-6" @submit.prevent="onSubmit">

      <div class="rounded-xl border bg-card p-5 space-y-4">
        <p class="text-sm font-medium text-muted-foreground uppercase tracking-wide">Personal info</p>

        <div class="space-y-2">
          <Label for="name">Name <span class="text-destructive">*</span></Label>
          <Input id="name" v-model="name" placeholder="Your name" />
          <p v-if="nameError" class="text-sm text-destructive">{{ nameError }}</p>
        </div>

        <div class="space-y-2">
          <Label for="dob">Date of birth</Label>
          <Input id="dob" v-model="dateOfBirth" type="date" />
        </div>

        <div class="space-y-2">
          <Label>Gender</Label>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="opt in GENDER_OPTIONS"
              :key="opt.value"
              type="button"
              class="h-10 rounded-lg border text-sm font-medium transition-all"
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

      <div class="rounded-xl border bg-card p-5 space-y-4">
        <p class="text-sm font-medium text-muted-foreground uppercase tracking-wide">Measurements</p>

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
          />
        </div>
      </div>

      <div class="flex gap-3">
        <Button
          type="button"
          variant="outline"
          class="px-5"
          @click="router.push({ name: 'profile' })"
        >
          Cancel
        </Button>
        <Button type="submit" class="flex-1" :disabled="updating">
          {{ updating ? "Saving..." : "Save changes" }}
        </Button>
      </div>

    </form>
  </div>
</template>
