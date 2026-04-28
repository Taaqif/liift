import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { userKeys } from "@/lib/queryKeys";

export type UserListItem = {
  id: number;
  username: string;
  email?: string;
  name: string;
  role: string;
};

export function useUsers() {
  const { data: users, isLoading } = useQuery({
    queryKey: userKeys.list(),
    queryFn: () => apiClient.get<UserListItem[]>("/users"),
  });
  return { users, isLoading };
}

export function useUpdateUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: number; role: string }) =>
      apiClient.put<UserListItem>(`/users/${id}/role`, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() });
    },
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => apiClient.delete(`/users/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() });
    },
  });
}
