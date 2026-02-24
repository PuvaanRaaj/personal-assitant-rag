import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/stores/auth';

const API_BASE_URL = '/api';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
      window.location.href = '/auth';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authApi = {
  register: async (email: string, password: string, name: string) => {
    const { data } = await api.post('/auth/register', { email, password, name });
    return data;
  },

  login: async (email: string, password: string) => {
    const { data } = await api.post('/auth/login', { email, password });
    return data;
  },

  refresh: async () => {
    const { data } = await api.post('/auth/refresh');
    return data;
  },
};

// Documents API
export const documentsApi = {
  list: async () => {
    const { data } = await api.get('/documents');
    return data;
  },

  upload: async (file: File, onProgress?: (progress: number) => void) => {
    const formData = new FormData();
    formData.append('file', file);

    const { data } = await api.post('/documents/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          onProgress(Math.round((progressEvent.loaded * 100) / progressEvent.total));
        }
      },
    });
    return data;
  },

  get: async (id: string) => {
    const { data } = await api.get(`/documents/${id}`);
    return data;
  },

  delete: async (id: string) => {
    await api.delete(`/documents/${id}`);
  },

  sync: async () => {
    const { data } = await api.post('/documents/sync');
    return data;
  },
};

// Query API
export const queryApi = {
  ask: async (question: string) => {
    const { data } = await api.post('/query', { question });
    return data;
  },
};
