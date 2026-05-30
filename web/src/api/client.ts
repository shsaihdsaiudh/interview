import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截：统一提取 data，类型安全
apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    console.error('API Error:', err);
    return Promise.reject(err);
  }
);

// 封装一层，自动从响应中提取 data
export async function apiGet<T>(url: string): Promise<T> {
  const res = await apiClient.get<T>(url);
  return res.data;
}

export default apiClient;
