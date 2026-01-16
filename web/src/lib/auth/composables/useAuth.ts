import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import {
  apiClient,
  type User,
  type LoginRequest,
  type RegisterRequest,
} from "@/lib/api";
import { decodeJWT } from "@/lib/jwt";

const user = ref<User | null>(null);
const loading = ref(false);

export function useAuth() {
  const router = useRouter();

  // Initialize auth state from localStorage
  const initAuth = () => {
    const token = apiClient.getToken();
    if (token) {
      const payload = decodeJWT(token);
      if (payload) {
        // Create user object from JWT payload
        user.value = {
          id: payload.user_id,
          username: payload.username,
        };
      } else {
        // Token is invalid or expired, remove it
        apiClient.setToken(null);
      }
    }
  };

  const login = async (credentials: LoginRequest) => {
    loading.value = true;
    try {
      const response = await apiClient.login(credentials);
      apiClient.setToken(response.token);
      user.value = response.user;
      router.push("/");
    } catch (err) {
      console.error(err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const register = async (data: RegisterRequest) => {
    loading.value = true;
    try {
      const response = await apiClient.register(data);
      apiClient.setToken(response.token);
      user.value = response.user;
      router.push("/");
    } catch (err) {
      console.error(err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const logout = () => {
    apiClient.setToken(null);
    user.value = null;
    router.push("/login");
  };

  const isAuthenticated = computed(() => {
    return !!apiClient.getToken() && !!user.value;
  });

  return {
    user: computed(() => user.value),
    loading: computed(() => loading.value),
    isAuthenticated,
    login,
    register,
    logout,
    initAuth,
  };
}
