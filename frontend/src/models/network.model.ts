export interface Network {
  // CamelCase properties (original)
  Name?: string;
  CIDR?: string;
  CreatedAt?: string;
  UpdatedAt?: string;
  
  // snake_case properties (from backend JSON)
  id?: string;
  cidr?: string;
  created_at?: string;
  updated_at?: string;
}