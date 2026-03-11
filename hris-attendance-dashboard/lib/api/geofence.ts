// src/lib/api/geofence.ts
import { apiClient } from './client'; // ✅ Fixed: Added missing slash
import type { Geofence, CreateGeofenceRequest, UpdateGeofenceRequest, CheckLocationResponse } from '@/types/geofence';

export const geofenceApi = {
  // Get all geofences
  getAll: async (): Promise<Geofence[]> => {
    const response = await apiClient.get('/geofences');
    return response.data.data;
  },

  // Get active geofences
  getActive: async (): Promise<Geofence[]> => {
    const response = await apiClient.get('/geofences/active');
    return response.data.data;
  },

  // Get geofence by ID
  getById: async (id: string): Promise<Geofence> => {
    const response = await apiClient.get(`/geofences/${id}`);
    return response.data.data;
  },

  // Create geofence
  create: async (data: CreateGeofenceRequest): Promise<Geofence> => {
    const response = await apiClient.post('/geofences', data);
    return response.data.data;
  },

  // Update geofence
  update: async (id: string, data: UpdateGeofenceRequest): Promise<Geofence> => {
    const response = await apiClient.put(`/geofences/${id}`, data);
    return response.data.data;
  },

  // Delete geofence
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/geofences/${id}`);
  },

  // Check location
  checkLocation: async (latitude: number, longitude: number, userId?: string): Promise<CheckLocationResponse> => {
    const response = await apiClient.post('/geofences/check', {
      latitude,
      longitude,
      user_id: userId,
    });
    return response.data.data;
  },
};