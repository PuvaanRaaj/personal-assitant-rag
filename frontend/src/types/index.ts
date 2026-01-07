export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Document {
  id: string;
  user_id: string;
  filename: string;
  file_type: string;
  file_size: number;
  file_hash: string;
  storage_path: string;
  total_chunks: number;
  created_at: string;
  updated_at: string;
}

export interface QueryResponse {
  answer: string;
  sources: Source[];
}

export interface Source {
  filename: string;
  page?: number;
  content?: string;
}

export interface ApiError {
  error: string;
  message?: string;
}
