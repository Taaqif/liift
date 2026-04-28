import { computed } from "vue";
import { useRouter } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type {
  User,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
} from "@/lib/auth/types";
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
    role: payload.role,
  };
}

export function useAuth() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: user, refetch: refetchUser } = useQuery({
    queryKey: authKeys.user(),
    queryFn: fetchUser,
    staleTime: 0, // Always consider data stale so it can be updated
    gcTime: 1000 * 60 * 10, // Keep in cache for 10 minutes
    retry: false,
    initialData: null,
  });

  const loginMutation = useMutation({
    mutationFn: (credentials: LoginRequest) => {
      return apiClient.post<AuthResponse>("/auth/login", credentials);
    },
    onSuccess: (response) => {
      apiClient.setToken(response.token);

      const payload = decodeJWT(response.token);
      queryClient.setQueryData<User | null>(authKeys.user(), {
        ...response.user,
        role: payload?.role,
      });
      if (!response.user.onboarding_complete) {
        router.push("/onboarding");
      } else {
        router.push("/");
      }
    },
    onError: (err) => {
      console.error(err);
    },
  });

  const registerMutation = useMutation({
    mutationFn: (data: RegisterRequest) => {
      return apiClient.post<AuthResponse>("/auth/register", data);
    },
    onSuccess: (response) => {
      apiClient.setToken(response.token);

      const payload = decodeJWT(response.token);
      queryClient.setQueryData<User | null>(authKeys.user(), {
        ...response.user,
        role: payload?.role,
      });
      router.push("/onboarding");
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
    return !!user.value;
  });

  const isAdmin = computed(() => user.value?.role === "admin");

  const loading = computed(() => {
    return loginMutation.isPending.value || registerMutation.isPending.value;
  });

  return {
    user: computed(() => user.value),
    loading,
    isAuthenticated,
    isAdmin,
    login,
    register,
    logout,
    initAuth,
  };
}
