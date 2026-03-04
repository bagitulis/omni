import { defineStore } from "pinia";
import { ref, computed } from "vue";

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(localStorage.getItem("authToken"));
  const user = ref<User | null>(
    localStorage.getItem("authUser")
      ? JSON.parse(localStorage.getItem("authUser") || "{}")
      : null
  );

  const isAuthenticated = computed(() => !!token.value && !!user.value);

  const setAuth = (auth: { token: string; user: User }) => {
    token.value = auth.token;
    user.value = auth.user;

    localStorage.setItem("authToken", auth.token);
    localStorage.setItem("authUser", JSON.stringify(auth.user));
  };

  const clearAuth = () => {
    token.value = null;
    user.value = null;

    localStorage.removeItem("authToken");
    localStorage.removeItem("authUser");
  };

  return {
    token,
    user,
    isAuthenticated,
    setAuth,
    clearAuth,
  };
});
