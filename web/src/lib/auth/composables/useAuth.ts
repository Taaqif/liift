import { computed } from "vue";
import { useRouter } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import {
  apiClient,
  type User,
  type LoginRequest,
  type RegisterRequest,
} from "@/lib/api";
import { decodeJWT } from "@/lib/jwt";
import { authKeys } from "@/lib/queryKeys";

async function fetchUser(): Promise<User | null> {
  const token = apiClient.getToken();
  if (!token) {
    return null;
  }

  const payload = decodeJWT(token);
  if (!payload) {
    apiClient.setToken(null);
    return null;
  }

  return {
    id: payload.user_id,
    username: payload.username,
  };
}

export function useAuth() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: user, refetch: refetchUser } = useQuery({
    queryKey: authKeys.user(),
    queryFn: fetchUser,
    staleTime: Infinity,
    gcTime: Infinity,
    retry: false,
    initialData: null,
  });

  const loginMutation = useMutation({
    mutationFn: (credentials: LoginRequest) => apiClient.login(credentials),
    onSuccess: (response) => {
      apiClient.setToken(response.token);

      queryClient.setQueryData<User | null>(authKeys.user(), response.user);
      router.push("/");
    },
    onError: (err) => {
      console.error(err);
    },
  });

  const registerMutation = useMutation({
    mutationFn: (data: RegisterRequest) => apiClient.register(data),
    onSuccess: (response) => {
      apiClient.setToken(response.token);

      queryClient.setQueryData<User | null>(authKeys.user(), response.user);
      router.push("/");
    },
    onError: (err) => {
      console.error(err);
    },
  });

  const login = async (credentials: LoginRequest) => {
    await loginMutation.mutateAsync(credentials);
  };

  const register = async (data: RegisterRequest) => {
    await registerMutation.mutateAsync(data);
  };

  const logout = () => {
    apiClient.setToken(null);

    queryClient.setQueryData<User | null>(authKeys.user(), null);
    router.push("/login");
  };

  const initAuth = () => {
    refetchUser();
  };

  const isAuthenticated = computed(() => {
    return !!apiClient.getToken() && !!user.value;
  });

  const loading = computed(() => {
    return loginMutation.isPending.value || registerMutation.isPending.value;
  });

  return {
    user: computed(() => user.value ?? null),
    loading,
    isAuthenticated,
    login,
    register,
    logout,
    initAuth,
  };
}
