import axios from 'axios';

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截：自动附带 JWT token
apiClient.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截：统一处理错误
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    // 401 / 404（用户不存在）时清除过期 token 和用户信息
    if (err.response?.status === 401) {
      removeToken();
      localStorage.removeItem(USER_KEY);
      window.dispatchEvent(new Event('auth-change'));
    }
    console.error('API Error:', err);
    return Promise.reject(err);
  }
);

// ── Token 管理 ──

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

// ── 封装请求方法 ──

// 封装一层，自动从响应中提取 data
export async function apiGet<T>(url: string): Promise<T> {
  const res = await apiClient.get<T>(url);
  return res.data;
}

export async function apiPost<T>(url: string, data?: unknown): Promise<T> {
  const res = await apiClient.post<T>(url, data);
  return res.data;
}

export async function apiPut<T>(url: string, data?: unknown): Promise<T> {
  const res = await apiClient.put<T>(url, data);
  return res.data;
}

export async function apiDelete<T>(url: string, data?: unknown): Promise<T> {
  const res = await apiClient.delete<T>(url, { data });
  return res.data;
}

export function getApiErrorMessage(err: unknown, fallback: string): string {
  return (
    (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
    fallback
  );
}

// ── 文件上传 ──

// apiUpload 上传文件（multipart/form-data），返回响应 data。
export async function apiUpload<T>(url: string, file: File, fieldName = 'avatar'): Promise<T> {
  const formData = new FormData();
  formData.append(fieldName, file);

  const token = getToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await axios.post<T>(`${apiClient.defaults.baseURL}${url}`, formData, {
    headers,
    timeout: 15000,
  });
  return res.data;
}

export default apiClient;
