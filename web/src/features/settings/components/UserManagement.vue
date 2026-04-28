<script setup lang="ts">
import { ref } from "vue";
import { toast } from "vue-sonner";
import { ShieldCheckIcon, UserIcon, TrashIcon } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useUsers, useUpdateUserRole, useDeleteUser } from "@/features/settings/composables/useUsers";
import { useAuth } from "@/lib/auth/composables/useAuth";

const { users, isLoading } = useUsers();
const { mutateAsync: updateRole, isPending: rolePending } = useUpdateUserRole();
const { mutateAsync: deleteUser, isPending: deletePending } = useDeleteUser();
const { user: currentUser } = useAuth();

const pendingId = ref<number | null>(null);

async function handleRoleChange(id: number, role: string) {
  pendingId.value = id;
  try {
    await updateRole({ id, role });
    toast.success("Role updated");
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Failed to update role");
  } finally {
    pendingId.value = null;
  }
}

async function handleDelete(id: number, username: string) {
  if (!confirm(`Delete user "${username}"? This cannot be undone.`)) return;
  pendingId.value = id;
  try {
    await deleteUser(id);
    toast.success(`User "${username}" deleted`);
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Failed to delete user");
  } finally {
    pendingId.value = null;
  }
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="isLoading" class="text-sm text-muted-foreground">Loading users...</div>
    <div v-else-if="!users?.length" class="text-sm text-muted-foreground">No users found.</div>
    <div v-else class="divide-y rounded-md border">
      <div
        v-for="u in users"
        :key="u.id"
        class="flex items-center gap-3 px-4 py-3"
      >
        <!-- Role icon -->
        <ShieldCheckIcon v-if="u.role === 'admin'" class="size-4 text-primary shrink-0" />
        <UserIcon v-else class="size-4 text-muted-foreground shrink-0" />

        <!-- Name + username -->
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium truncate">
            {{ u.name || u.username }}
            <span v-if="u.id === currentUser?.id" class="text-xs text-muted-foreground ml-1">(you)</span>
          </p>
          <p v-if="u.name" class="text-xs text-muted-foreground truncate">@{{ u.username }}</p>
        </div>

        <!-- Role selector -->
        <Select
          :model-value="u.role"
          :disabled="u.id === currentUser?.id || (rolePending && pendingId === u.id)"
          @update:model-value="handleRoleChange(u.id, $event as string)"
        >
          <SelectTrigger class="w-28 h-8 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="admin">Admin</SelectItem>
            <SelectItem value="user">User</SelectItem>
          </SelectContent>
        </Select>

        <!-- Delete -->
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 text-muted-foreground hover:text-destructive"
          :disabled="u.id === currentUser?.id || (deletePending && pendingId === u.id)"
          @click="handleDelete(u.id, u.username)"
        >
          <TrashIcon class="size-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
