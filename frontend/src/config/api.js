import axios from 'axios';
import { auth } from './firebase';

const API_BASE_URL = "https://api.shrten.live";

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for logging and auth
api.interceptors.request.use(async (config) => {
  console.log(`[API Request] ${config.method?.toUpperCase()} ${config.url}`);
  
  if (auth && auth.currentUser) {
    try {
      const token = await auth.currentUser.getIdToken();
      config.headers.Authorization = `Bearer ${token}`;
      console.log('[API Request] Auth token added');
    } catch (error) {
      console.warn('Failed to get auth token:', error);
    }
  } else {
    console.warn('[API Request] No authenticated user');
  }
  
  return config;
}, (error) => {
  console.error('[API Request Error]', error);
  return Promise.reject(error);
});

// Add response interceptor for logging
api.interceptors.response.use(
  (response) => {
    console.log(`[API Response] ${response.config.method?.toUpperCase()} ${response.config.url} - Status: ${response.status}`);
    return response;
  },
  (error) => {
    if (error.response) {
      console.error(`[API Error] ${error.config?.method?.toUpperCase()} ${error.config?.url} - Status: ${error.response.status}`, error.response.data);
    } else if (error.request) {
      console.error('[API Error] No response received:', error.message);
    } else {
      console.error('[API Error]', error.message);
    }
    return Promise.reject(error);
  }
);

export { API_BASE_URL };
export default api;
