<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { useAuth } from "@/lib/auth/composables/useAuth";
import { useProfile } from "@/features/profile/composables/useProfile";
import { Button } from "@/components/ui/button";
import { Pencil, LogOut, User, Calendar, Ruler, Weight, ChevronRight } from "lucide-vue-next";

const router = useRouter();
const { logout } = useAuth();
const { profile, loading } = useProfile();

const initials = computed(() => {
  if (profile.value?.name) {
    return profile.value.name
      .split(" ")
      .map((w) => w[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  }
  return (profile.value?.username?.[0] ?? "?").toUpperCase();
});

const displayName = computed(() => profile.value?.name || profile.value?.username || "");

const formattedDob = computed(() => {
  if (!profile.value?.date_of_birth) return null;
  const d = new Date(profile.value.date_of_birth + "T00:00:00");
  return d.toLocaleDateString(undefined, { year: "numeric", month: "long", day: "numeric" });
});

const genderLabel = computed(() => {
  const map: Record<string, string> = { male: "Male", female: "Female", other: "Other" };
  return profile.value?.gender ? (map[profile.value.gender] ?? profile.value.gender) : null;
});

const infoRows = computed(() => {
  const rows: { icon: any; label: string; value: string }[] = [];
  if (formattedDob.value) rows.push({ icon: Calendar, label: "Date of birth", value: formattedDob.value });
  if (genderLabel.value) rows.push({ icon: User, label: "Gender", value: genderLabel.value });
  if (profile.value?.height_cm) rows.push({ icon: Ruler, label: "Height", value: `${profile.value.height_cm} cm` });
  if (profile.value?.weight_kg) rows.push({ icon: Weight, label: "Weight", value: `${profile.value.weight_kg} kg` });
  return rows;
});
</script>

<template>
  <div class="pb-10 max-w-xl">
    <div class="mb-8 flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">Profile</h1>
      <Button variant="outline" size="sm" @click="router.push({ name: 'profile-edit' })">
        <Pencil class="h-4 w-4 mr-2" />
        Edit
      </Button>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-4">
      <div class="rounded-2xl border bg-card p-6 space-y-4">
        <div class="flex items-center gap-4">
          <div class="w-16 h-16 rounded-full bg-muted animate-pulse" />
          <div class="space-y-2">
            <div class="h-5 w-32 bg-muted animate-pulse rounded" />
            <div class="h-4 w-24 bg-muted animate-pulse rounded" />
          </div>
        </div>
      </div>
    </div>

    <template v-else-if="profile">
      <!-- Avatar + name card -->
      <div class="rounded-2xl border bg-gradient-to-br from-primary/8 via-primary/4 to-transparent p-6 mb-4">
        <div class="flex items-center gap-4">
          <div class="w-16 h-16 rounded-full bg-primary/15 border border-primary/20 flex items-center justify-center shrink-0">
            <span class="text-xl font-bold text-primary">{{ initials }}</span>
          </div>
          <div class="min-w-0">
            <p class="text-xl font-bold truncate">{{ displayName }}</p>
            <p class="text-sm text-muted-foreground truncate">@{{ profile.username }}</p>
            <p v-if="profile.email" class="text-sm text-muted-foreground truncate">{{ profile.email }}</p>
          </div>
        </div>
      </div>

      <!-- Info rows -->
      <div v-if="infoRows.length > 0" class="rounded-xl border bg-card overflow-hidden mb-4">
        <div
          v-for="(row, i) in infoRows"
          :key="row.label"
          class="flex items-center gap-3 px-4 py-3.5"
          :class="{ 'border-b': i < infoRows.length - 1 }"
        >
          <component :is="row.icon" class="size-4 text-muted-foreground shrink-0" />
          <span class="text-sm text-muted-foreground flex-1">{{ row.label }}</span>
          <span class="text-sm font-medium">{{ row.value }}</span>
        </div>
      </div>

      <!-- Empty state: no profile info yet -->
      <div v-else class="rounded-xl border border-dashed bg-card p-6 text-center mb-4">
        <p class="text-sm text-muted-foreground mb-3">No profile details yet.</p>
        <Button variant="outline" size="sm" @click="router.push({ name: 'profile-edit' })">
          Complete your profile
        </Button>
      </div>

      <!-- Account actions -->
      <div class="rounded-xl border bg-card overflow-hidden">
        <button
          class="w-full flex items-center gap-3 px-4 py-3.5 border-b hover:bg-muted/40 transition-colors text-left"
          @click="router.push({ name: 'profile-edit' })"
        >
          <Pencil class="size-4 text-muted-foreground shrink-0" />
          <span class="text-sm flex-1">Edit profile</span>
          <ChevronRight class="size-4 text-muted-foreground" />
        </button>
        <button
          class="w-full flex items-center gap-3 px-4 py-3.5 hover:bg-muted/40 transition-colors text-left text-destructive"
          @click="logout"
        >
          <LogOut class="size-4 shrink-0" />
          <span class="text-sm flex-1">Log out</span>
        </button>
      </div>
    </template>
  </div>
</template>
