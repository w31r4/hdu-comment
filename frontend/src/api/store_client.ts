import axios from 'axios';
import type { Store, Review, PaginatedResponse } from '../types';

const api = axios.create({
  baseURL: '/api/v1'
});

export interface StoreSearchParams {
  query?: string;
  page?: number;
  page_size?: number;
}

export interface CreateStoreWithReviewInput {
  store_name: string;
  store_address: string;
  store_phone?: string;
  store_category?: string;
  store_description?: string;
  review_title: string;
  review_content: string;
  rating: number;
}

export interface SubmitReviewInput {
  store_id: string;
  content: string;
  rating: number;
}

export interface UpdateReviewInput {
  content: string;
  rating: number;
}

// 搜索店铺
export const searchStores = async (params: StoreSearchParams = {}): Promise<PaginatedResponse<Store>> => {
  const { data } = await api.get<PaginatedResponse<Store>>('/stores', { params });
  return data;
};

// 获取店铺详情
export const getStore = async (storeId: string): Promise<Store> => {
  const { data } = await api.get<Store>(`/stores/${storeId}`);
  return data;
};

// 获取用户对店铺的评价
export const getMyStoreReview = async (storeId: string, token: string): Promise<Review | null> => {
  try {
    const { data } = await api.get<Review>(`/stores/${storeId}/my-review`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    return data;
  } catch (error: any) {
    if (error.response?.status === 204) {
      return null; // 没有评价
    }
    throw error;
  }
};

// 创建新店铺并提交评价
export const createStoreWithReview = async (input: CreateStoreWithReviewInput, token: string): Promise<{ store: Store; review: Review }> => {
  const { data } = await api.post<{ store: Store; review: Review }>('/stores/with-review', input, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return data;
};

// 提交店铺评价
export const submitStoreReview = async (input: SubmitReviewInput, token: string): Promise<Review> => {
  const { data } = await api.post<Review>('/reviews/store', input, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return data;
};

// 更新店铺评价
export const updateStoreReview = async (reviewId: string, input: UpdateReviewInput, token: string): Promise<Review> => {
  const { data } = await api.put<Review>(`/reviews/store/${reviewId}`, input, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return data;
};

// 获取店铺的所有评价
export const getStoreReviews = async (storeId: string, params: { page?: number; page_size?: number } = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>(`/stores/${storeId}/reviews`, { params });
  return data;
};