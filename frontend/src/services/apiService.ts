import axios from 'axios';
import { User } from '../models/user.model';

// Create an instance of axios
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL,
});

// Request interceptor for adding the JWT token to requests
api.interceptors.request.use((config) => {
  const user: User | null = JSON.parse(localStorage.getItem('user')!);
  if (user && user.accessToken) {
    config.headers!.Authorization = `Bearer ${user.accessToken}`;
  }
  return config;
}, error => Promise.reject(error));

// Response interceptor for handling errors globally
api.interceptors.response.use((response) => response, (error) => {
  if ([401, 403].includes(error.response.status)) {
    // Logout logic here or emit an event to logout
  }
  return Promise.reject(error);
});

export default api;
