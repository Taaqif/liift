const API_BASE_URL = "/api";

export interface ApiError {
  error: string;
}

class ApiClient {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem("auth_token", token);
    } else {
      localStorage.removeItem("auth_token");
    }
  }

  getToken(): string | null {
    if (!this.token) {
      this.token = localStorage.getItem("auth_token");
    }
    return this.token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const token = this.getToken();
    const isFormData = options.body instanceof FormData;
    const headers: HeadersInit = {
      Authorization: token ? `Bearer ${token}` : "",
      ...(isFormData ? {} : { "Content-Type": "application/json" }),
      ...options.headers,
    };

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      // Handle 401 Unauthorized - token is invalid or expired
      if (response.status === 401) {
        this.setToken(null);
      }

      const error: ApiError = await response.json().catch(() => ({
        error: "An error occurred",
      }));
      throw new Error(error.error || `HTTP error! status: ${response.status}`);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const contentType = response.headers.get("content-type");

    if (!contentType || !contentType.includes("application/json")) {
      await response.text().catch(() => {});
      return undefined as T;
    }

    const text = await response.text();
    if (!text || text.trim() === "") {
      return undefined as T;
    }

    return JSON.parse(text);
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "GET" });
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    const isFormData = data instanceof FormData;
    return this.request<T>(endpoint, {
      method: "POST",
      body: data ? (isFormData ? data : JSON.stringify(data)) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    const isFormData = data instanceof FormData;
    return this.request<T>(endpoint, {
      method: "PUT",
      body: data ? (isFormData ? data : JSON.stringify(data)) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "DELETE" });
  }
}

export const apiClient = new ApiClient();

// Cache for blob URLs to avoid refetching the same image
const blobUrlCache = new Map<string, string>();

/**
 * Fetches an image with authentication and returns a blob URL.
 * This avoids exposing the access token in the URL.
 *
 * @param imagePath - The image path (can be relative or absolute URL)
 * @returns A blob URL that can be used in <img> tags, or undefined if no image path
 */
export async function getImageUrl(
  imagePath: string | null | undefined,
): Promise<string | undefined> {
  if (!imagePath) {
    return undefined;
  }

  // If it's already a full external URL, return as-is
  if (
    imagePath.startsWith("http") &&
    !imagePath.startsWith(window.location.origin)
  ) {
    return imagePath;
  }

  // Check cache first
  if (blobUrlCache.has(imagePath)) {
    return blobUrlCache.get(imagePath);
  }

  try {
    // Extract the API endpoint path
    let endpoint = imagePath;
    if (imagePath.startsWith(window.location.origin)) {
      endpoint = imagePath.replace(window.location.origin, "");
    }
    // Ensure it starts with /api
    if (!endpoint.startsWith(API_BASE_URL)) {
      endpoint = `${API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;
    }

    // Fetch the image with authentication
    const token = apiClient.getToken();
    const response = await fetch(`${window.location.origin}${endpoint}`, {
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });

    if (!response.ok) {
      if (response.status === 401) {
        apiClient.setToken(null);
      }
      return undefined;
    }

    // Convert to blob and create object URL
    const blob = await response.blob();
    const blobUrl = URL.createObjectURL(blob);

    // Cache the blob URL
    blobUrlCache.set(imagePath, blobUrl);

    return blobUrl;
  } catch (error) {
    console.error("Failed to fetch image:", error);
    return undefined;
  }
}

/**
 * Revokes a blob URL and removes it from the cache.
 * Call this when you're done with an image to free up memory.
 */
export function revokeImageUrl(imagePath: string | null | undefined): void {
  if (!imagePath) {
    return;
  }

  const blobUrl = blobUrlCache.get(imagePath);
  if (blobUrl) {
    URL.revokeObjectURL(blobUrl);
    blobUrlCache.delete(imagePath);
  }
}
