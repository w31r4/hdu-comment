import axios, { AxiosError, AxiosRequestConfig } from 'axios';
import type {
  AuthResponse,
  PaginatedResponse,
  Review,
  User,
  Store,
  CreateStoreRequest,
  CreateReviewRequest,
  UpdateReviewRequest,
  CreateReviewForNewStoreRequest,
  AutoCreateReviewResponse
} from '../types';

const api = axios.create({
  baseURL: '/api/v1'
});

const rawApi = axios.create({
  baseURL: '/api/v1'
});

let accessToken: string | null = null;
let refreshToken: string | null = null;
let refreshExecutor: (() => Promise<AuthResponse | null>) | null = null;
let refreshPromise: Promise<AuthResponse | null> | null = null;

const setAuthorizationHeader = (config: AxiosRequestConfig, token: string | null) => {
  if (!config.headers) config.headers = {};
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  } else {
    delete config.headers.Authorization;
  }
};

api.interceptors.request.use((config) => {
  if (accessToken) {
    setAuthorizationHeader(config, accessToken);
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const responseStatus = error.response?.status;
    const originalRequest: any = error.config;

    if (responseStatus === 401 && refreshToken && refreshExecutor && !originalRequest?._retry) {
      originalRequest._retry = true;

      if (!refreshPromise) {
        refreshPromise = refreshExecutor().finally(() => {
          refreshPromise = null;
        });
      }

      const result = await refreshPromise;
      if (result && result.access_token) {
        accessToken = result.access_token;
        refreshToken = result.refresh_token;
        setAuthorizationHeader(originalRequest, accessToken);
        return api(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);

export const setAuthTokens = (access: string | null, refresh: string | null) => {
  accessToken = access;
  refreshToken = refresh;
};

export const clearAuthTokens = () => {
  accessToken = null;
  refreshToken = null;
};

export const setRefreshExecutor = (executor: (() => Promise<AuthResponse | null>) | null) => {
  refreshExecutor = executor;
};

export const getRefreshToken = () => refreshToken;

// --- Generic Query Params ---

export interface ListQueryParams {
  page?: number;
  page_size?: number;
  query?: string;
  sort?: string; // e.g., 'created_at', '-rating'
  order?: 'asc' | 'desc';
  category?: string;
  status?: string;
}

// --- Auth API ---

export const register = async (email: string, password: string, displayName: string): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>('/auth/register', {
    email,
    password,
    display_name: displayName
  });
  return data;
};

export const login = async (email: string, password: string): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>('/auth/login', { email, password });
  return data;
};

export const refreshTokens = async (token: string): Promise<AuthResponse> => {
  const { data } = await rawApi.post<AuthResponse>('/auth/refresh', { refresh_token: token });
  return data;
};

export const logout = async (token: string): Promise<void> => {
  await api.post('/auth/logout', { refresh_token: token });
};

// --- User API ---

export const fetchMe = async (): Promise<User> => {
  const { data } = await api.get<User>('/users/me');
  return data;
};

export const fetchMyReviews = async (params: ListQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/users/me/reviews', { params });
  return data;
};

// --- Store API ---

export const searchStores = async (params: ListQueryParams = {}): Promise<PaginatedResponse<Store>> => {
  const { data } = await api.get<PaginatedResponse<Store>>('/stores', { params });
  return data;
};

export const getStore = async (storeId: string): Promise<Store> => {
  const { data } = await api.get<Store>(`/stores/${storeId}`);
  return data;
};

export const createStore = async (payload: CreateStoreRequest): Promise<Store> => {
  const { data } = await api.post<Store>('/stores', payload);
  return data;
};

// --- Review API ---

export const fetchReviews = async (params: ListQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/reviews', { params });
  return data;
};

export const getStoreReviews = async (
  storeId: string,
  params: ListQueryParams = {}
): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>(`/stores/${storeId}/reviews`, { params });
  return data;
};

export const fetchReviewDetail = async (id: string): Promise<Review> => {
  const { data } = await api.get<Review>(`/reviews/${id}`);
  return data;
};

export const createReviewForNewStore = async (
  payload: CreateReviewForNewStoreRequest
): Promise<AutoCreateReviewResponse> => {
  const { data } = await api.post<AutoCreateReviewResponse>('/reviews', payload);
  return data;
};

export const createReviewForStore = async (storeId: string, payload: CreateReviewRequest): Promise<Review> => {
  const { data } = await api.post<Review>(`/stores/${storeId}/reviews`, payload);
  return data;
};

export const updateReview = async (storeId: string, reviewId: string, payload: UpdateReviewRequest): Promise<Review> => {
  const { data } = await api.patch<Review>(`/stores/${storeId}/reviews/${reviewId}`, payload);
  return data;
};

export const deleteReview = async (storeId: string, reviewId: string): Promise<void> => {
  await api.delete(`/stores/${storeId}/reviews/${reviewId}`);
};

export const uploadReviewImage = async (id: string, file: File): Promise<void> => {
  const formData = new FormData();
  formData.append('file', file);
  await api.post(`/reviews/${id}/images`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
};

// --- Admin API ---

export const fetchPendingReviews = async (params: ListQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/admin/reviews/pending', { params });
  return data;
};

export const updateReviewStatus = async (id: string, status: 'approved' | 'rejected', reason?: string): Promise<Review> => {
  const { data } = await api.put<Review>(`/admin/reviews/${id}/status`, { status, reason });
  return data;
};

export const adminDeleteReview = async (id: string): Promise<void> => {
  await api.delete(`/admin/reviews/${id}`);
};

export const fetchPendingStores = async (params: ListQueryParams = {}): Promise<PaginatedResponse<Store>> => {
  const { data } = await api.get<PaginatedResponse<Store>>('/admin/stores/pending', { params });
  return data;
};

export const adminCreateStore = async (payload: CreateStoreRequest): Promise<Store> => {
  const { data } = await api.post<Store>('/admin/stores', payload);
  return data;
};

export const updateStoreStatus = async (id: string, status: 'approved' | 'rejected', reason?: string): Promise<Store> => {
  const { data } = await api.put<Store>(`/admin/stores/${id}/status`, { status, reason });
  return data;
};

export const adminDeleteStore = async (id: string): Promise<void> => {
  await api.delete(`/admin/stores/${id}`);
};
