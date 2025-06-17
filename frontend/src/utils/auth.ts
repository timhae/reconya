// src/utils/auth.ts
import { API_BASE_URL } from '../config';
export const checkAuth = async (): Promise<boolean> => {
  const token = localStorage.getItem('token');
  if (!token) return false;

  try {
    const response = await fetch(`${API_BASE_URL}/check-auth`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    return response.ok;
  } catch (error) {
    console.error('Error checking authentication:', error);
    return false;
  }
};
