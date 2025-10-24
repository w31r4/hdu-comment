export interface User {
  email: string;
  display_name: string;
  role: 'user' | 'admin';
  created_at?: string;
  updated_at?: string;
}

export interface Store {
  id: string;
  name: string;
  address: string;
  phone?: string;
  category?: string;
  description?: string;
  average_rating: number;
  total_reviews: number;
  created_at: string;
}

export interface Image {
  id: string;
  url: string;
}

export interface Author {
  id: string;
  display_name: string;
}

export interface Review {
  id: string;
  author: Author;
  title: string;
  content: string;
  rating: number;
  images?: Image[];
  created_at: string;
  updated_at: string;
  store?: Store;
  status?: 'approved' | 'pending' | 'rejected';
  rejection_reason?: string;
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

// --- Request Payloads ---

export interface CreateStoreRequest {
  name: string;
  address: string;
  phone?: string;
  category?: string;
  description?: string;
}

export interface CreateReviewRequest {
  title: string;
  content: string;
  rating: number;
}

export interface UpdateReviewRequest {
  title?: string;
  content?: string;
  rating?: number;
}

export interface StoreInfoForReview {
  name: string;
  address: string;
}

export interface CreateReviewForNewStoreRequest extends CreateReviewRequest {
  store: StoreInfoForReview;
}

// --- Special API Responses ---

export interface AutoCreateReviewResponse {
  store: Store;
  review: Review;
  is_new_store: boolean;
}
