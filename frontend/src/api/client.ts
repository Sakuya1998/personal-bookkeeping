import axios from 'axios';
import { useAppStore } from '../store/appStore';

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
});

client.interceptors.request.use((config) => {
  const token = useAppStore.getState().token;
  if (token) {
    config.headers = config.headers ?? {};
    (config.headers as Record<string, string>).Authorization = `Bearer ${token}`;
  }
  return config;
});

let lastUnauthorizedAt = 0;

client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      useAppStore.getState().logout();

      const now = Date.now();
      if (now - lastUnauthorizedAt > 1000) {
        lastUnauthorizedAt = now;
        window.dispatchEvent(
          new CustomEvent('auth:unauthorized', {
            detail: { next: window.location.pathname + window.location.search },
          }),
        );
      }
    }
    return Promise.reject(err);
  },
);

export function resetUnauthorizedHandlingForTests() {
  lastUnauthorizedAt = 0;
}

export default client;
