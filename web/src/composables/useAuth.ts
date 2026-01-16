import { ref, computed, nextTick } from "vue";
import { useRouter } from "vue-router";
import {
  apiClient,
  type User,
  type LoginRequest,
  type RegisterRequest,
  type AuthResponse,
} from "@/lib/api";
import { decodeJWT } from "@/lib/jwt";

const user = ref<User | null>(null);
const loading = ref(false);

export function useAuth() {
  const router = useRouter();

  // Set user details from auth response
  const setUserFromResponse = (response: AuthResponse) => {
    apiClient.setToken(response.token);
    user.value = {
      id: response.user.id,
      username: response.user.username,
      email: response.user.email,
      created_at: response.user.created_at,
      updated_at: response.user.updated_at,
    };
  };

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
      setUserFromResponse(response);
      await nextTick();
      await router.push("/");
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
      setUserFromResponse(response);
      await nextTick();
      await router.push("/");
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
    // Only check user.value since it's reactive and gets set when token is valid
    return !!user.value;
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
