import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { profileKeys } from "@/lib/queryKeys";
import type { Profile, UpdateProfilePayload } from "@/features/profile/types";

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (payload: UpdateProfilePayload) =>
      apiClient.put<Profile>("/users/me", payload),
    onSuccess: (data) => {
      queryClient.setQueryData(profileKeys.me(), data);
    },
  });

  return {
    updateProfile: mutation.mutateAsync,
    updating: mutation.isPending,
    error: mutation.error,
  };
}
