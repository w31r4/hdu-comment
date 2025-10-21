export interface User {
  id: string;
  email: string;
  display_name: string;
  role: 'user' | 'admin';
  created_at?: string;
}

export interface Store {
  id: string;
  name: string;
  address: string;
  phone?: string;
  category?: string;
  description?: string;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  average_rating: number;
  total_reviews: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ReviewImage {
  id: string;
  review_id: string;
  storage_key: string;
  url: string;
  created_at: string;
}

export interface Review {
  id: string;
  store_id: string;
  store?: Store;
  title: string;
  content: string;
  rating: number;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  author_id: string;
  author?: User;
  images?: ReviewImage[];
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationMeta;
}
